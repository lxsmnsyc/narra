# Log: lsp-hover-parity

- Branch `fix/lsp-hover-parity` fast-forwarded to `feat/batch-20` (0fd1e813).
- Base narrac: `sh selfhost/bootstrap.sh <dir>` on 0fd1e813.
- Fixtures: `testdata/lsp/hover_names` (aliases, struct literal, `Shape as S`, `area as measure`, pattern, builtins) and `testdata/lsp/hover_errors` (package with a type error).
- Proof: expected.txt recorded with the new narrac. `NARRAC=<base narrac> sh scripts/test.sh --suite lsp --filter hover_` fails on `hover_names` (module alias in a template answers the template `string`, no module hover for `Arrays`, `Shapes` in a pattern, `S`, the `as` local, the struct fields; `Point` instead of `struct Point`; `Shape` instead of the variant). Passes on the branch.
- `hover_errors` passes on base and branch: the missing `uses` line of a private function in a package with errors does not change (see outcome).
- Builtins: the golden records `print`, `Int.toString` and `Arrays.length` (this one has no signature, so it keeps the call's type `int`).
- `narrac query type-of` goldens in `testdata/query/basic` changed in two lines (variant call, module alias in a template): same `findTypeAt`.
