# Log: lsp-hover-parity

- Branch `fix/lsp-hover-parity` fast-forwarded to `feat/batch-20` (0fd1e813).
- Base narrac: `sh selfhost/bootstrap.sh <dir>` on 0fd1e813.
- Fixtures: `testdata/lsp/hover_names` (aliases, struct literal, `Shape as S`, `area as measure`, pattern, builtins) and `testdata/lsp/hover_errors` (package with a type error).
- Proof: expected.txt recorded with the new narrac. `NARRAC=<base narrac> sh scripts/test.sh --suite lsp --filter hover_` fails on `hover_names` (module alias in a template answers the template `string`, no module hover for `Arrays`, `Shapes` in a pattern, `S`, the `as` local, the struct fields; `Point` instead of `struct Point`; `Shape` instead of the variant). Passes on the branch.
- `hover_errors` passes on base and branch: the missing `uses` line of a private function in a package with errors does not change (see outcome).
- Builtins: the golden records `print`, `Int.toString` and `Arrays.length` (this one has no signature, so it keeps the call's type `int`).
- `narrac query type-of` goldens in `testdata/query/basic` changed in two lines (variant call, module alias in a template): same `findTypeAt`.
- Review fixes: `findAliasUsePath` returns the module path of an alias only when the identifier does not resolve to a local; `describeModuleUse` and `collectImportHovers` both use it (the old `aliases` map is gone). `describeStructLiteral` appends ` as <written name>` when the written name differs. Removed the three function-describing comments.
- Proof: new case `testdata/lsp/hover_shadow` (`const Shapes = 1; return Shapes` with `import * as Shapes`, and `P { ... }` for `Point as P`). `narrac query type-of` on `Shapes` printed `module app/shapes` on cec154c4 and `int` now; the struct literal now prints `struct Point as P`.
- `collectImportHovers` and `describeModuleUse` still each call `describeMember` once; the alias check, the part that was wrong, is shared.
