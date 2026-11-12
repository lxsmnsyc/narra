# Outcome: fmt-std-sources

## Built

- `selfhost/tests/src/sources.nra` checks `std/src`; `std/src/coroutine.nra` and `std/src/option.nra` formatted; `selfhost/src/stdtext.nra` regenerated.
- `endsWithBlock` in `selfhost/src/fmtexprs.nra` checks every operand of a same-precedence chain.
- Golden `testdata/fmt/expr_block_operand_middle.{nra,fmt}`.

## Proof

- base: feat/batch-10, `narrac fmt --check std/src` lists coroutine.nra and option.nra; `x + {...} + y` in an assignment prints the block and `y` one indent deeper than `x`.
- branch: `narrac fmt --check std/src selfhost/src selfhost/tests` lists nothing; the chain prints all operands level (golden passes).

## Deviations

- None. No files in selfhost/src needed reformatting. oxfmt was not available, so the layout follows its rules by reasoning and the existing expr_block_operand_level golden.

## Stubs left

- None.

## Gotchas

- A block as the first operand now also gets the level layout.
- The seed is not regenerated (orchestrator, end of batch).

## Proposed learnings

- None.

## Verification

- `--suite sources`, `--suite fmt` (86), `--suite generated` (17) pass; `sh scripts/test.sh`: PASS, 7111 cases.

## Size

- Estimated S, actual S.
