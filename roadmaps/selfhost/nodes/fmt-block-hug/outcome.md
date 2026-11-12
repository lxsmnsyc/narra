# Outcome: fmt-block-hug

## Built

- `selfhost/src/fmtexprs.nra`: `isPlainBlock`, added to `isHuggable`, so a plain block in the last call argument stays on the call line (`addAll(mutable xs, {`). The hug is skipped when the last argument has an attached comment or the argument before it has one (a second pass would move the comment). The same guard fixes a last-argument arrow (`(x) => { // c`), which had the bug on the base.
- Binary operand: no hug (oxfmt never hugs it). When a chain in an assignment ends in a block, `printAssigned` (`fmt.nra`) prints it with `printLevelChain` (new, `fmtexprs.nra`): the block is level with the first operand (`x.n +` then `{` at the same indent), as oxfmt prints operands after `=`. A `return` keeps the old indent: oxfmt wraps a return operand chain in parentheses, and fmt does not add parentheses.
- fmt goldens: `expr_block_hug_argument`, `expr_block_hug_fallback_w40`, `expr_block_operand_level`, `comment_block_hug_argument`.
- Separate final commit `fmt-block-hug: reformat sources`: 9 files (`fmtexprs.nra`; testdata `arg_eval_order`, `block_expressions`, `named_args`; three `testdata/errors` samples and two `.err` goldens).

## Proof

- base (feat/batch-9, 9b546355, compiler built from the unchanged fmt files): `narrac fmt --stdin` on `const total = addAll(mutable xs, { ... })` prints `addAll(`, `mutable xs,`, the block as its own argument, `},`, `)`. On `testdata/errors/mutable_read_conflict_block_operand.nra` it prints `const total =`, then `x.n +`, then the block indented two further levels. oxfmt keeps a last-argument function or object on the call line (not run here: not installed).
- branch: `addAll(mutable xs, {` ... `})`; and `const total =` / `x.n +` / `{` ... `}` with `x.n +` and `{` at one indent. The fallback case still breaks every argument. A second pass changes nothing (fmt suite stable check).
- Review (oxfmt 0.66.0): oxfmt never hugs a block-like operand and keeps operands level; both now match.

## Deviations

- Touched testdata error samples and their `.err` goldens: the reformat moved the reported positions.
- None beyond touching `fmt.nra` and the testdata error positions.
- When the block is not the last operand (`const mid = x + {...} + y`), oxfmt prints all operands level, but Narra keeps the old extra indent, because `endsWithBlock` (`fmtexprs.nra`) checks only the rightmost operand.

## Stubs left

- None.

## Gotchas

- Formatting moves lines in `testdata/errors` samples: update the `// error:` header and rerun `sh scripts/test.sh --suite errors --update`.
- `std/src/coroutine.nra` and `option.nra` are not formatted on the base either; the sources suite does not check `std/src`.
- A compiler built without `-DNARRA_RELEASE` exits 103 on the copy check; use the bootstrap output for suites.

## Proposed learnings

- None.

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite fmt --filter expr_block_hug`, `--suite sources`, `--suite errors`: pass.
- Full `sh scripts/test.sh`: PASS, 7070 cases in 247s (fixed point reached; the sources suite passes on the reformatted tree).

## Size

- Estimated M, actual S.

Landed on feat/batch-9 as ca50df1e.
