# Outcome: shared-write-sweep

## Built

- **E0978 in the checker modules: 141 to 0.** Every file in `selfhost/src` except the ones shared-write-sweep-rest owns (`emit*`, `cnames`, `fmt*`, `query*`, `lsp*`, `trivia*`, `parser`, `lexer`, `modules`, `manifest`, `driver`). The edits:
  - The fixes E0978 offers (`const x = copy(checker.<path>)` and similar), about 130 sites over several rounds.
  - Loop-held writes (22 loops): the loop runs over `Arrays.values(copy(...))`. In release builds `copy(x)` is one retain (`emitCopyCall`, `stdcopy.nra`), so each costs one retain per loop, not a copy of the list.
  - Values stored in the checker from a parameter or a declaration get `copy(...)` where they are stored: `declaration: copy(implDecl)`, `copy(funcDecl)`, `copy(constDecl)`, `typeArguments: copy(typeArguments)`, `environment: copy(environment)`, `arguments: copy(arguments)`, `pos: copy(pos)`, `namePos: copy(namePos)`, `copy(aliasDecl.pos|typeParameters|body)`.
  - Reads moved before the write: `collectExpressionAndArms` (exprs.nra) reads the arm's parts before it recurses; `checkBlockBody` (stmts.nra) restores `loops` and `labels` from `finished` (nothing writes those fields of a pushed block); `checkFor` computes `isContentsType` before writing the local; `scopes.nra` copies `stdModule.functions`, `calls.nra` copies `funcParams` and `provider.typeArguments`, `effects.nra` copies the callee it found.
  - The effect pass records `copy(expression)` in `state.valueLists` (one retain per recorded value, only in bodies with a function store). The earlier `recordValue` helper, which relied on the rule not following `mutable` arguments of user functions, is gone.
  - Save/restore sites (O9): `resolveAlias` saves `checker.environment` after `makeParameters`, which does not read it (no copy); `checkArrowParts` clears `loops`, `labels` and `blocks` before `pushArrow`/`pushScope`, which do not read them (no copy); `checkReferenceFor` keeps `copy(checker.loopBorrows)`, because the push needs the outer list and the new one at once and Narra has no pop (the list is as long as the loop nesting).
- **E0976 in the checker modules (switch removed locally): 104 to 0.**
  - `from <parameter>` on every checker function whose result shares a parameter, enum results included (`Type`, `Option<...>`, `Expr`). Where a result shares a second parameter, the return is `copy(...)`.
  - The Walker trait functions in check.nra carry the `from` of their one implementation (body.nra): `checkExpression`, `checkPatterns`, `checkMapLiteral`, `checkSpread` and `checkStdCall` are `from want`, `checkMapIndex` is `from mapValue`, `checkTuple`, `checkCall`, `checkConstructorName`, `checkBlockExpression`, `checkBlockBody` and `checkArrow` are `from checker`.
  - Results of the all-scalar structs `Pos` and `CLayout` get `copy(...)` at the return instead of `from` (`getExpressionPos`, `getPatternPos`, `findStatementPos`, `declPos`, `findFromNamePos`, the layout readers): the copy emits the value itself and callers do not start to share.
- **`copy(x)` takes enums with a tracked payload** (orchestrator decision, option a). `canShareType` moved from `sharedresults.nra` to `stdcopy.nra` (exported); `checkCopyCall` accepts every type it accepts. `isCopyTracked` keeps its meaning (a write changes the value in place), because the shared-write walk uses it for exactly that. E0975 still rejects an enum whose payloads hold only scalars and strings (`copy of Option<int> copies nothing`).
  - **Debug copy check.** A retain of a heap enum counts only its payload box, so a write to an array taken out of a copied recursive enum was reported as unexplained (exit 103). The emitter (`emitdecls.nra`: `creditFunction`, `creditCode`, `creditPartCode`, `holdsHeapEnum`) writes a `credit_<enum>` function for each heap enum, inside `#ifndef NARRA_RELEASE`, that gives a credit to every counted value in the payload; `emitCopyCall` calls it after its retain, under `if (narra_copy_on)`, when the copied type holds a heap enum. Release builds are unchanged.
  - **The fixes.** The E0976 copy fix (`buildCopyFix`, sharedresults.nra) is offered only when `copy` takes the value's type; the E0978 copy fix (sharedwrites.nra) uses the same test, so a held enum now gets a fix.
  - Samples: `testdata/copy_enum.nra` (+ `.out`; an `Option<Array<int>>` and a recursive enum with a struct payload, run under the copy check), `testdata/diagnostics/from_enum_copy_fix.nra` (the E0976 copy fix on an `Option<Array<int>>` result checks clean), `testdata/errors/copy_untracked` now uses `Option<int>`.
  - Docs: D94 amendment in DECISIONS.md (with option b rejected and why); SPEC.md section 6, Copies on shared values: the rule, a sample with an `Option<Array<int>>` and an enum with a struct payload, and the E0975 sample for `Option<int>`.
- **Seed regenerated** once (commit e61449e6), after the copy commit, so the compiler can write `copy(...)` of a `Type`. The runtime text did not change, so one update was enough. The seed is older than the final source again (expected).
- `selfhost/narra.toml` keeps `shared-results = "unchecked"`.

## Proof

- Counts, `narrac check selfhost` (E0978) and the same with the switch removed locally (E0976):

| | base e8f0b61f | branch |
|---|---|---|
| E0978, checker modules | 141 | 0 |
| E0976, checker modules | 104 | 0 |
| E0978, rest modules | 56 | 62 |
| E0976, rest modules | 102 | 103 |

- `narrac emit selfhost`, instructions retired, interleaved runs on a loaded machine: base (compiler built from feat/batch-6 aaee3950, on its own source) 2,360,696,798 / 2,386,594,621 / 2,361,282,884, median 2,361,282,884; branch (85fb381b, on its own source) 2,424,407,448 / 2,388,295,737 / 2,386,229,084, median 2,388,295,737: +1.1%. The first pass alone (before the enum marks) measured +0.18%; the rest comes from the copies at enum returns (`checkStdCall` and the checks it dispatches to are `from want` and copy the checker-derived types they return) and the stored-parameter copies, each one retain.
- No behavior change: the branch compiler and the base compiler emit the same C for the base source (aaee3950 `selfhost/`) after preprocessing with `-DNARRA_RELEASE`; the raw C differs only by added `credit_*` functions and their calls inside `#ifndef NARRA_RELEASE`. The branch's own emitted C is 182,642 lines; stage 1 and stage 2 emit identical C.

## Deviations

- **Seed update inside the node.** GOAL.md has the last node of a batch regenerate the seed. The enum marks need `copy` of a `Type` in the compiler, which the old seed rejects, so this node regenerated it once (e61449e6). The batch's own end-of-batch seed updates still apply.
- **The emitter changed** (`emitdecls.nra`, `stdcopy.nra`'s `emitCopyCall`), which belong to shared-write-sweep-rest's files for the sweep. The edits are the copy feature, not sweep edits; the rest modules' warnings are untouched.
- `narrac apply-fix` was not run per fix. One `narrac check selfhost --json` per round and a script applied the first fix of every checker-module diagnostic (the same edits `apply-fix` writes), deduplicating equal edits and leaving conflicting ones for the next round.
- New warnings in the rest modules, from the new marks: E0978 56 to 62 (five `Effects.checkEffects(mutable checker)` sites in driver.nra, emitter.nra, lspcache.nra, lspnavigation.nra and query.nra, from `declarePackage(package): Checker from package`, and `argumentCode(mutable emitter, ...)` in emitexprs.nra, from `variantFieldType(...): Type from emitter` in stdemithelp.nra), E0976 102 to 103. All have fixes; shared-write-sweep-rest owns them.

## Stubs left

- None. shared-write-sweep-rest fixes the rest modules and removes the switch.

## Gotchas

- `copy(x)` is a retain in release builds, so loop-over-copy and stored-parameter copies are cheap. The real copy happens later, on the first write to a shared value, as before.
- A `credit_<enum>` call walks the whole payload of a copied heap enum, recursively, only when `NARRA_CHECK_COPIES` is set. A copy of a very deep enum in a debug run with the check on can use a lot of C stack.
- `from` on a widely used function moves warnings into its callers, including callers in other modules. Mark from the leaves up and recheck after each round.
- The Walker trait's `from` markers are fixed by its one implementation. A second implementation must return values that fit them (`copy(...)` otherwise).
- `canShareType` lives in `stdcopy.nra` now; import it from there.

## Proposed learnings

- Return an all-scalar struct taken from a parameter as `copy(...)`, not with `from` | the copy is free and callers do not start to share | selfhost/src | no
- Store a parameter or declaration into the checker as `copy(...)` | otherwise every later `mutable checker` call warns for the caller's value | selfhost/src | yes (E0978)
- A new kind of value that `copy` accepts needs its parts reachable by the debug copy check's credits | otherwise copies on write of those parts are reported as unexplained | selfhost/src/emitdecls.nra | yes (a `copy_*` sample)

## Verification

- `narrac check selfhost`: 0 E0978 in the checker modules; with the switch removed, 0 E0976 in them.
- `sh selfhost/bootstrap.sh out/new2` on 85fb381b: stage 1 and stage 2 emit identical C (182,642 lines); the seed is older than the source (expected).
- Base source emitted by the base and the branch compilers: same release C (above).
- `NARRAC=<branch narrac> sh scripts/test.sh --suite errors --suite habits --suite codes --suite sources --suite dumps --suite fmt --suite spec --suite query --suite generated`: PASS, 1407 cases. `--suite programs --filter copy_`: PASS (with `copy_enum`, debug and release).
- `JOBS=3 sh scripts/test.sh` on 85fb381b: PASS in 370 s, 6327 cases, 0 failed, with the bootstrap job (stage 1 and stage 2 identical).

## Size

- Estimated L, actual L, plus the copy-of-enum feature and one seed update after the orchestrator's decision.
