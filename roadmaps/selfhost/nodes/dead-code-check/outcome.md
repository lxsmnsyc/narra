# Outcome: dead-code-check

## Built

- `selfhost/src/deadcode.nra`: `checkDeadCode(mutable checker, block)`. For each block it finds the first statement that always leaves (return, break, continue, a call to panic or Process.exit, an `if` and `else` that both leave, a block or `match` whose paths all leave, a `while (true)` or `for (;;)` with no leaving `break`) and reports E1500 at the next statement. Only the first dead statement is reported. The fix deletes from that statement to the closing brace.
- One call in `Stmts.checkBlock` and one for the `if` then-branch (`selfhost/src/stmts.nra`, which checks its statements without `checkBlock`). Function bodies, closures, loops, arms, else-branches and nested blocks all go through one of the two. When the dead code is on lines of its own, the fix removes whole lines.
- `pathends.nra` now exports `findWhileOpenPath` and `findCForOpenPath`.
- `diagnostics.nra`: E1500 row, explanation, range note.
- Block expressions (`endsInReturn` in `stmts.nra`) now count a `while (true)` loop with no leaving `break` as ending in a return; before, such code needed a dead `return` after the loop.
- Swept the dead statements it found: selfhost/src, std/src (json, toml, unicode, async__net; `stdtext.nra` regenerated), selfhost/tests/src/codes.nra, testdata samples.
- SPEC.md rule next to "Unused names are errors"; guide chapter 3 has a checked example.
- Tests: `testdata/errors/dead_code_*` (return, panic, break, continue, if/else, match, loop).

## Proof

Not a fix node.

## Deviations

- Edited outside the owned paths: `selfhost/src/pathends.nra`, `selfhost/src/stmts.nra`, `selfhost/src/*.nra` sweep (shared), `std/src/{json,toml,unicode,async__net}.nra` and `selfhost/src/stdtext.nra` (generated), `selfhost/tests/src/codes.nra`, other `testdata/` samples and goldens (block_expressions_break_*, match_block_values_break_across, const_eval_steps, typesdump err_loops, async_*, recursive, block_expressions).
- Severity is an error, per the maintainer decision of 2026-11-14.

## Stubs left

- None.

## Gotchas

- `narrac apply-fix` prints the file instead of writing it, fixes one diagnostic per run, and may include unrelated fixes (it dropped `copy(...)` calls and a table row). Read the diff after using it.
- Dead code in std modules that `selfhost` does not import only shows up in the guide and program suites.
- `match` arms with an expression body count as leaving only for panic or Process.exit calls.
- The `seed` is not regenerated; the batch's seed update must run after this lands, and the seed rule still applies (the compiler source does not use any new feature).

## Proposed learnings

- A statement after one that always leaves is E1500; do not write a trailing `return` after `while (true)` | the check rejects it | selfhost/src, std/src | yes (the sources build)

## Verification

- `sh scripts/test.sh`: see final line in the report.

## Size

- Estimated M, actual M. The sweep touched many files but each change is one line.
