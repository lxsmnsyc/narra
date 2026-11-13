# Outcome: arrow-arg-order

## Built

- `ClosureInfo.writesCaptured` (`selfhost/src/check.nra`), set by `markCapturedWrite` (`selfhost/src/funcvalues.nra`) for every arrow that captures a `let` the arrow, or one nested in it, assigns or changes the contents of. Called from `markChangedCapture` (content writes) and from `checkAssignment` in `selfhost/src/stmts.nra` (whole assignment of a `let`).
- `mayWrite` in `selfhost/src/emitexprs.nra` counts an arrow with `writesCaptured` as a writer, so earlier arguments are held.
- Sample `testdata/arg_order_arrow_writes.nra` and `.out`: `1,6` for the writing arrow, `2,10` for a block-bodied arrow that only reads.

## Proof

- base: af35fdcb, `narrac run testdata/arg_order_arrow_writes.nra` printed `6,6` then `2,10`.
- branch: same command prints `1,6` then `2,10`.

## Deviations

- Edited `selfhost/src/check.nra`, `selfhost/src/funcvalues.nra`, `selfhost/src/stmts.nra`, which are not in the brief's paths. The emitter has no record of which arrows write a captured `let` (`refCaptures` holds only `const`s), so the checker must record it.

## Stubs left

- None. The seed is not regenerated.

## Gotchas

- async-core: an async arrow or block that writes a captured `let` goes through the same checker marks, so check that `markCapturedWrite` runs for its body.

## Proposed learnings

- None.

## Verification

- `narrac emit selfhost` output: 215832 lines with the base compiler and with the branch compiler (no growth).
- `sh scripts/test.sh`: see log.md.

## Size

- Estimated S, actual S.
