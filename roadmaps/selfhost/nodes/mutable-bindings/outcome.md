# Outcome: mutable-bindings

## Built

D81 in `selfhost/src`. The D85 loop form is not built; node mutable-for-of builds it.

- **Locals.** `LocalInfo.binding` (`BindingKind`: Plain, Pattern, Loop, Destructured, `check.nra`) splits the old `isMutable` meaning. `isMutable` now means only "the name can be reassigned" (`let`). `canChangeContents(local)` means "fields and elements can be assigned and the local passed as `mutable`": a `mutable` parameter, or a Plain `const`/`let`. A Loop local also keeps `loopSource` (the array place, as text) and `loopContents` (the element is a contents type) for the hint. `setBinding` marks the declaration that the parser's `for` desugaring makes from the hidden `__for_L_C` name as a Loop binding.
- **Rule 1.** `checkMutableType` (`check.nra`), called from `resolveSignature` (`resolve.nra`) and `resolveTraitDecl` (`traits.nra`): a `mutable` parameter must be a struct (not a tuple), `Array`, `Map`, `Box`, `StringBuilder` or `Generator`/`Coroutine` (`isContentsType`). E0948, also for a bare type parameter. `Self` is allowed in a trait declaration.
- **Rule 2.** `checkAssignment` (`stmts.nra`): assigning a whole `mutable` parameter is E0949.
- **Rules 3 and 4.** `checkContentsChange` (`exprs.nra`), called for field and element writes (`checkAssignment`) and `mutable` arguments (`checkMutableArgument`). A local `const` passes. A plain parameter is E0227 (write) or E0315 (`mutable` argument), with new texts. A pattern, loop or destructured binding is E0950. Its hint for `for (const p of players)` over an array of a contents type is `for (mutable p of players)` (the D85 form, not built yet), and `players[index].hp` for a destructured loop binding. Top-level constants keep E0290 (text now "copy it into a local first, as in `const copy = NAME`").
- **Rule 5.** `checkMutableSelf` (`traits.nra`): an `implement` of a trait with a `mutable Self` function for a type with no contents is E0952. A bare `mutable value: T` is E0948. `Dynamic<Trait>` is not built, so its rule has nothing to check (README).
- **Rule 6.** `markChangedCapture` (`funcvalues.nra`): a change to a captured local `const` adds it to `ClosureInfo.refCaptures` of every closure between the change and the declaration. A heap closure in that chain is E0951. `checkBorrowed` (E0614) treats these captures like a `let`. The emitter (`isCapturedByReference`, `emitclosures.nra`) stores them as `T*` in the environment, as for a `let`. `findOutermostCapture` is the chain search, shared with `useLocal`.
- **Rule 7.** E0319, E0929 and E0609 were never gated on `let`; the rewritten tests now pass `const` locals.
- **selfhost/src rewrites.** `patterns.nra` `findMissingCases` takes `mutable budget: CaseBudget` (new struct, caller `match.nra`). `declsdump.nra` `dumpModule` returns the next implement index and `dumpImpl` takes it by value. `modules.nra` `sortStrings` sorts in place. The callers keep `let`, because the seed requires a `let` for a `mutable` argument. Bootstrap fixed point holds.
- **Tests.** New: `testdata/mutable_contents.nra` and `.out` (a `const` struct changed and passed as `mutable`, a `const` array, map and box changed, a copy that stays a copy, nested stack closures that push to a captured `const`, a `const` only read by a closure). New errors (`testdata/errors/mutable_contents_*.nra`): `type`, `whole`, `param`, `param_arg`, `for`, `for_entries`, `pattern`, `destructure`, `generic`, `implement`, `escaping`, `borrow`. Rewritten to struct counters or array fields, each keeping its rule: `closure_borrow`, `match_block_values_exclusive`, `mutable_read_ok`, `ternary_mutable`, `unused_mutable_output` (run), 15 errors tests (`mutable_read_conflict_*`, `mutable_twice`, `ternary_mutable_conflict`, `match_block_values_mutable_conflict`, `closures_*`, `closure_borrow_mutable_param`), the 4 `let_narrowing_same_statement_*` (the narrowing now ends at a stack closure that assigns the `let`), `diagnostics/mutable_read_conflict_fix.nra`, `typesdump/bodies/err_statements.nra`, `typesdump/closures/err_closures.nra`. Goldens updated: the three `const_*` errors, `err_assign.dump`, `err_arrays_maps.dump` (its `const` push line is gone).
- **Removed tests.** `errors/box_assign_const` (valid now; in the run sample), `errors/let_narrowing_mutable_arg` (after the rewrite it duplicated `closure_borrow_narrowing`), `errors/let_narrowing_same_statement_param` and `errors/mutable_read_conflict_option` (both rested on a `mutable Option` parameter; the second's rule is in `mutable_read_conflict_nested`).
- **Docs.** `docs/guide` chapter 2 (`const` fixes the name; the deep-`const` sample became a reassignment error), chapter 3 (closure capture sentence, escaping closures cannot change captures), chapter 9 (plain parameter hint, `const` locals in every sample, a section on contents types and no whole assignment with an E0948 sample, starting kit, closures section, the `int` reporter sample on a struct, "Freeing memory"), chapter 11 (`const` is shallow with a run sample, copies and read-only copying bindings, `addOne` on a struct, the comparison list, the plain-parameter sample). `compiler/README.md`: the D81 "not built" line is replaced by the unbuilt parts (Dynamic, FixedArray, identity types), the loop-form line points at mutable-for-of, the unused-names line says "only written through". `SPEC.md`: dropped two "proposed" marks on D81 rules 8 and 9.

## New codes

- E0948: `mutable` parameter of a type with no contents (also a bare type parameter).
- E0949: whole assignment to a `mutable` parameter.
- E0950: change through a pattern, loop or destructured binding (5 formats).
- E0951: a closure that is not a stack closure changes a captured value.
- E0952: `implement` for a type with no contents of a trait with a `mutable Self` function.
- E0953: reserved, unused (would be `Dynamic<Trait>` passed to a `mutable Self` function, once `Dynamic` exists).
- Changed formats: E0227 (new row for plain parameters), E0315 (now plain parameters only), E0290.

## Gate decisions

1. **The bare generic case shares E0948** with the non-contents types, with "found type parameter T". Alternative: its own code. Why: it is rule 1 for a type the compiler cannot prove has contents; it left a code free for rule 6.
2. **Whole assignment of a `const`, a pattern or a loop binding keeps E0227 with its old text** ("is constant; declare it with `let` to change it"). Alternative: reword for each binding kind. Why: the text is still right for a whole assignment, and it keeps those goldens.
3. **The `for (const ...)` hint already names `for (mutable p of players)`**, as the orchestrator asked, when the source is an array place whose element has contents and the binding is the loop's own name. For a destructured loop binding (`#[index, player]`), the hint is `players[index].hp`. Alternative: always the index hint until D85 lands. Risk: until mutable-for-of lands, the suggested form is a parse error. The guide shows no sample of it.
4. **A closure that only reads a `const` another closure changes stays by value.** `refCaptures` is per closure, so only the closures on the chain of a change hold the `const` by pointer (D81 rule 9 exactly).
5. **Narrowing tests whose trigger was a `mutable Option` argument** were rewritten to a stack closure that assigns the `let`, or removed when that made them duplicates. Alternative: drop all of them. Why: the "same statement" end is still a rule, reached now only through closures.
6. **The guide does not show the read-only loop binding error**, since its hint names the unbuilt loop form. Chapter 11 describes the rule in prose. mutable-for-of can add the sample with the form.

## What this means for mutable-for-of

- Mark the loop binding with `binding = BindingKind.Plain`-like rights but no reassignment: the simplest is a new `BindingKind.LoopReference`, accepted by `canChangeContents` and rejected for whole assignment with E0949 (rule 3 of D85).
- `checkFor` (`stmts.nra`) already finds the source place (`Arrays.values`/`entries` argument, or the iterable) and its element type; reuse that and `isContentsType`.
- The emitter's closure capture by reference keys on `isMutable` or `refCaptures`; a loop reference binding is already a pointer, so `isReference` (`emitstate.nra`) must return true for it, as it does for `param && isMutable`.
- After D85 lands, the hint text "write `for (mutable p of players)`" is correct as it stands.

## Deviations

- The brief's "34 fixtures and 28 whole assignments" were counted in the retired `compiler/testdata`; the live tree is `testdata/`, and the rewrite covered every test the new rules broke there.
- Three error tests removed instead of rewritten (see Removed tests).
- `SPEC.md` got a two-word edit (dropping "proposed"); `DECISIONS.md` still says D81 is not built (not in this node's shared paths).

## Stubs left

- A `distinct type` over a struct or `Array` is not a contents type to `isContentsType`, so `mutable value: Wrapped` is E0948. D81 does not name distinct types; revisit if a program needs it.
- E0953 for `Dynamic<Trait>`, until `Dynamic` exists.

## Gotchas

- The seed still runs D27 rules: `selfhost/src` cannot pass a `const` as `mutable` or write a field of a `const` until the seed is regenerated.
- `--update` of the errors suite does not check `// error:` lines; fix their positions by hand after a rewrite.
- `LocalInfo.binding` has no default (an enum default is not a constant to the checker), so each `LocalInfo { ... }` literal names it (`declareLocal`, `createPatternLocal`).

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh out`: stage 2 and stage 3 emit identical C. Stage 1 is the committed seed compiling the changed `selfhost/src`, so the seed accepts every rewrite.
- `NARRAC=out/narrac sh scripts/test.sh`: PASS in 439 s (every shard and the bootstrap job; machine shared with other workers).
- After the rebase onto origin/feat/batch-1 (27ceacc, with std/json, std/shape, sized numbers, f32 and the `lsp` suite): bootstrap fixed point holds, `NARRAC=$PWD/out/narrac JOBS=2 sh scripts/test.sh` PASS in 176 s (4989 tests and the bootstrap job), and the seed build `JOBS=2 sh scripts/test.sh` PASS in 163 s. The landed code needed no D81 fix.
- Docker: not run.

## Size

- Estimated L, actual M to L (the checker change was smaller than estimated; the test and guide rewrite was the bulk).
