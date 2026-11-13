# Outcome: bound-trait-shared-results

## Built

- `findSharedArgument` (`selfhost/src/sharedresults.nra`) finds the trait of a call through a bound `<T: Trait>` from the callee name (`lookupTrait`) and reads the `from` marker of the trait function, as it does for a Dynamic.
- `testdata/errors/shared_result_bound_write.nra`: error `Arrays.push(mutable got, ...) may change got`. `testdata/shared_result_bound_copy.nra`: the same code with `copy(...)`, runs clean and prints 3.

## Proof

- base: af35fdcb (feat/batch-12), `narrac build -o bin main.nra && NARRA_CHECK_COPIES=1 ./bin` on the bound-call program: compiles, "copies on write that no copy(...) explains: array 1", exit 103.
- branch: the same program fails to compile with `may change got, which item still holds for the caller`; with `copy(...)` it runs and prints 3.

## Deviations

- The check now finds results the compiler's own sources shared without a marker. Fixed in files outside the brief: `std/src/shape.nra` (`from value` on 10 generic Shape helpers), `selfhost/src/exprs.nra`, `match.nra`, `calls.nra`, `stdcopy.nra` (markers and `copy(...)` on 7 sites), `deriveshape.nra` (generated Shape functions get `from __value`), `testdata/shape.nra` (hand-written Shape impls get `from value`), and the regenerated `selfhost/src/stdtext.nra`.
- The run sample `testdata/shared_result_bound_copy.{nra,out}` is outside `testdata/errors/`.
- No new error code.

## Stubs left

- None.

## Gotchas

- The seed is not updated; stage 1 built from the seed compiles the new sources.
- A change to `std/src/shape.nra` needs `sh scripts/test.sh --suite generated --update` for `stdtext.nra`.
- `exprs.nra` is a hot file; the edits are one-line markers and `copy(...)` wraps, so a rebase keeps both sides.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C.
- `sh scripts/test.sh --filter shared_result_bound`: PASS, 10 cases.
- `sh scripts/test.sh`: PASS, 8113 cases.

## Size

- Estimated S, actual S.
