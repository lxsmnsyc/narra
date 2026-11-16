# Outcome: query-hovers

## Built

- `narrac query [--package DIR] [--json] hovers FILE` in `selfhost/src/tools/query.nra`. It lists every named site of the file (declarations, parameters, locals, fields, callees, pattern bindings) with the text of `findTypeAt`, the same function the LSP hover calls, plus the `uses` line the LSP adds. It also lists the errors reported in the file.
- Contract `contracts/query-hovers.md`, frozen at version 1. Line and column are 1-based and count code points, and `length` counts the same.
- Works on a file with check errors (effects are left out) and on a package that fails to load (empty hovers, the file's errors).
- Names the site index does not cover: the module alias in `import * as M` and at each use (`module std/int`), the imported member after it (its signature with effects), named imports, the enum name before a variant (`enum Shape`), and a variant in a pattern or call with its fields (`Shape.Circle(radius: int)`). A name with nothing to show is left out, never a placeholder. Bindings in match, `is`, destructuring and `for` patterns come from the site index.
- Cases: `testdata/query/hovers` (a clean file, a generic module, a file that does not exist, text mode, a non-ASCII string before a name) and `testdata/query/hovers_error`, and `testdata/query/hovers_names` for imports, match and `is` patterns, destructuring and `for`.
- `docs/guide/12-tools.md` and `SPEC.md` list the command. The seed is regenerated.

## Proof

Not a fix node. One fix on the way: a hover on a call of a private function in a package with errors panicked in `findRefEffects` (the effect pass had not run, `functionKeys` is empty). `describeCallable` now skips the effects when the pass did not run, which also fixes the LSP hover there.

## Result

Not an experiment.

## Deviations

- Diagnostic `length` is the length of the token that starts at the position, or 1: the checker records only a start position.
- The `uses` line of a hover is `uses alloc, io`; a builtin call such as `print(...)` answers for the call (`void` plus effects), as `type-of` does.

## Stubs left

- None.

## Gotchas

- Columns are code points. A consumer that indexes JavaScript strings converts them to UTF-16 units.
- `sh scripts/test.sh --update` rewrites goldens with the seed's compiler, so run `scripts/seed-update.sh` before it when the change adds a command.

## Proposed learnings

- Regenerate the seed before `--update` for a node that adds a command | the update runs the old compiler | roadmaps/selfhost/CLAUDE.md | no

## Verification

- `sh scripts/seed-update.sh` twice: the second prints "seed unchanged".
- `sh scripts/test.sh`: see the node report.

## Size

- Estimated M, actual M.
