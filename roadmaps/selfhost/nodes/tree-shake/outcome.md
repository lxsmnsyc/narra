# Outcome: tree-shake

## Built

- Emission-driven tree shaking of functions, generic instances, trait implementations, function bodies and lazy constants, for user and std modules alike. Types, composite helpers (rc, credit/uncredit, map functions) and runtime text are still emitted whole (tree-shake-types).
- `Emitter.references: Array<string>` and `noteReference(mutable emitter, name)` (`emitstate.nra`): every site that writes the C name of a function or constant getter notes it. The sites are `emitFunctionCall` and `constRef` (`emitexprs.nra`), the function-value thunk in `emitFuncValue` (`emitclosures.nra`), and the user implementation that a trait function (eq, hash, cmp, show) forwards to in `writeDefinition` (`emittraits.nra`).
- `emitReachableBodies(mutable emitter): ReachableBodies` (`emit.nra`): the roots are the entry function (`emitter.entry`, the name `main` calls) or, in a test build, every test block. It walks `emitter.references` with a cursor and emits each newly named function (`Consts.emitFunction`) or lazy constant getter (`Consts.emitGetter`). Then it writes the trait functions, and repeats while a drain after the trait pass emits a new body. Closures, generators and instances need no extra edge: their bodies are emitted while the owner is emitted, and their calls are noted there. A folded build-condition branch is never emitted, so it adds no edge.
- `emitconsts.nra`: `lazyConstants` and `emitGetter` (now returns its text) are exported; `emitConstPrototypes(emitter, constants)` takes the constant list; `emitConstBodies` became `emitConstRelease(emitter, constants)` (the release function only).
- Output order is the old order, filtered: prototypes and bodies follow `Checker.programFunctions`, getters and the release function follow `lazyConstants`. What changes with emission order is the numbering of string literals and the order of closure and trait-function text. These follow discovery order, which is deterministic.
- Sample `testdata/tree_shake.nra` (+ `.out`) and two emit cases in `selfhost/tests/src/emitchecks.nra`: `tree-shake/unreached` (no unreached exported function, no `pick<string>` instance, no unreached `Show` implementation, no unreached constant; the reached ones are present) and `tree-shake/std-source` (`bench/fib` has no `compare`, `divmod`, `floorDiv`).

Where the reachable set lives, for tree-shake-types: `ReachableBodies` (exported struct in `emit.nra`), computed inside `emitProgram`. `functions: Array<bool>` is parallel to `checker.programFunctions` (after `dropDuplicateTypes`), and `functionTexts` holds the text, or "" when unreached. `constants: Array<int>` lists the reached lazy constants (`Checker.consts` indexes) in `lazyConstants` order, and `getterTexts` their getters. Static constants are inlined at each use, so a reached body carries their references. The set is known only after the bodies are emitted. emitProgram already emits bodies before it writes the constant and function prototypes. It still writes the type sections (comment table, typedefs, rc functions, map functions) before the bodies, so tree-shake-types must move those after `emitReachableBodies`, or collect the types a body uses during emission in the same way (a note at the `cachedCType` or `cachedTypeId` sites). `emitter.references` keeps every noted name in note order, duplicates included.

## Proof

- base: feat/batch-7 at 2598f249 (after contracts and return-trailer), compiler from `sh selfhost/bootstrap.sh` built in a clean copy of that tree, Apple M2, shared machine (load 40 to 60).
  - `narrac emit bench/fib/fib.nra`: 1086 lines, with 3 ext__int bodies (`compare`, `divmod`, `floorDiv`).
  - `narrac emit selfhost`: 183665 lines, 8500463 bytes, 4591 `static ... {` definitions.
  - `/usr/bin/time -l narrac emit selfhost`: 2526.8M / 2540.3M / 2518.4M instructions (median 2.527G), peak 141.1 to 141.5 MB.
- branch: same commands, run right after, with the compiler from the branch's bootstrap.
  - fib: 1028 lines, 0 ext__int bodies.
  - selfhost: 172394 lines (-11271, -6.1%), 7928503 bytes (-572 KB, -6.7%), 4409 definitions (-182).
  - Instructions 2479.1M / 2487.2M / 2486.1M (median 2.486G, -1.6%), peak 131.6 to 132.3 MB (-9.8 MB). The body texts go straight into the C text at the end, with no intermediate bodies builder. Before that change (first version, on 1a8a90c0) peak was 3 to 5 MB above base.
- First version, on 1a8a90c0: 185157 to 173647 lines; median 2.498G to 2.461G instructions.
- Folded branch probe: an `if (Build.os == Os.Windows)` holding an arrow that calls `onlyWindows` and a call to `wrap<string>`. The base C has both; the branch C has neither the function, the instance nor the closure.

## Deviations

- Files outside `shared:`, kept to the edge records: `emitstate.nra` (the field and `noteReference`), `emitexprs.nra` (2 notes), `emitclosures.nra` (1 note), `emittraits.nra` (1 note), `emitconsts.nra` (constants split into per-getter emission and a filtered release function, so constants can be shaken).
- `selfhost/tests/src/emitchecks.nra`: the two emit cases. `testdata/tree_shake.nra`/`.out`: the sample (the programs suite runs it under the sanitizers and the leak check).
- `emitdecls.nra` is unchanged: credit/uncredit functions are per-type helpers, left to tree-shake-types as the orchestrator said.
- Unsupported-construct markers in unreachable code are no longer reported (that code is not emitted), so `narrac emit` can exit 0 where it exited 3. No test depended on that.

## Stubs left

- Types, typedefs, rc/credit functions, map functions and runtime text of unreached code: tree-shake-types.

## Gotchas

- A new site that writes a function or getter name must call `noteReference`, or the C fails to compile or link with an undefined static function. The full suite and the bootstrap catch it.
- The trait pass notes its implementation each time it runs, so the loop's stop test is "a drain emitted no new body", not "no new reference". The first version tested the reference count and hung on `testdata/params_pointer_calls.nra` (user `==`).
- `noteReference(mutable emitter, ...)` trips the shared-write check if a value borrowed from `emitter.checker` is used after it; note after the last use, or `copy(emitter.checker)` first.
- Unsupported-construct markers in unreachable code are no longer reported, since that code is not emitted (accepted by the orchestrator).
- The trait functions are written once per drain round that emitted something after them; selfhost needs one round.

## Proposed learnings

- Compute what to emit from the names the emitter writes (note at the naming site), not from a second AST walk | no edge kind can be missed, and folded branches drop out on their own | selfhost/src/emit*.nra | yes (bootstrap and the C compile fail on a missed name)

## Verification

- `sh selfhost/bootstrap.sh <dir>` after the rebase onto 2598f249: fixed point (stage 2 = stage 3, 172394 lines); seed older than source, as expected mid-batch. The emit suite (both tree-shake cases) passes with that compiler.
- `NARRAC=<stage1> sh scripts/test.sh --suite <s>` for programs, emit, sources, leaks, panics, packages, edge, runner, guide, spec: pass.
- `sh scripts/test.sh` after the rebase: PASS in 345 s (8 shards, 820 to 821 cases each, and the bootstrap job).

## Size

- Estimated M, actual M.

Landed on feat/batch-7 as cc31389d.
