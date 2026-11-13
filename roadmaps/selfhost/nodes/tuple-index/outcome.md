# Outcome: tuple-index

## Built

- `pair[0]` reads a tuple element with that element's type. `checkTupleIndex` in `selfhost/src/exprs.nra`; the emitter case is in `emitIndexRead` (`selfhost/src/emitexprs.nra`) and writes `(target).f_N`.
- E1010 (index not a constant integer) and E1011 (index out of range) are rows and explanations in `selfhost/src/diagnostics.nra`.
- Tuple elements are not places: `findAssignedLocal` returns -1 for an index on a tuple, so `pair[0] = 2`, `+=` and `++` report E0226.
- Samples: `testdata/tuple_index.nra` and `.out`. Errors: `testdata/errors/tuple_index_*.nra` and `.err` (let, parameter, range, negative, assign, compound, increment, mutable).
- SPEC.md Tuples bullet, a note on D53 in DECISIONS.md, README line removed.

## Proof

Not a fix node.

## Result

Not an experiment.

## Deviations

- `selfhost/src/check.nra` (not in the brief's paths): a new body table `tupleIndexes: Map<int, int>`, because the emitter cannot evaluate a `const` index itself.
- `stmts.nra`, `sharedwrites.nra`, `sharedresults.nra` and the eight other `Expr.IndexExpr` readers (operands, effects, emitstmts, emitmaps, stdcopy, closures, contracts, querysites) were not edited; none needed a tuple case.
- `update(mutable pair[0])` reports the existing mutable-argument error, not E0226.
- The feature is not used in selfhost/src or std/src; the seed is not regenerated.

## Stubs left

- None.

## Gotchas

- The emitter reads tuple indexes from `tupleIndexes`. Code that builds an `IndexExpr` on a tuple without going through `checkIndex` has no entry and gets "unsupported: index".

## Proposed learnings

- None.

## Verification

- `NARRAC=<built narrac> sh scripts/test.sh`: `test.sh: PASS, 7615 cases in 222s`.
- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C.
- Sample built with `-sanitize` and run with `NARRA_CHECK_COPIES=1`: no report; the programs suite runs it clean.

## Size

- Estimated S, actual S.
