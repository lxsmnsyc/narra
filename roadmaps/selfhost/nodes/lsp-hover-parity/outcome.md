# Outcome: lsp-hover-parity

## Built

- `findTypeAt` and `describeSite` in `selfhost/src/tools/query.nra` now answer what `narrac query hovers` shows, so `narrac lsp` gets it:
  - a module alias (also inside template strings and patterns) shows `module PATH`; `Alias.member` shows the member;
  - a struct literal shows `struct Name` at the type name and `field: type` at each written field;
  - a variant shows `Enum.Variant(fields)`;
  - the enum name before a variant shows `enum Shape`, and `enum Shape as S` where the source wrote an alias (the declared name stays, the written name follows);
  - the local name in `import { area as measure }` shows the imported name's text.
- New helpers: `findDetachedHover`, `findQualifierHovers`, `describeModuleUse`, `collectImportNames`, `findStructLiteralIndex`, `describeStructLiteral`, `describeStructField`, `findTokenIndex`, `findModuleAliasPath`.
- Cases `lsp/hover_names` and `lsp/hover_errors` (`testdata/lsp`).

## Proof

- base: 0fd1e813, `NARRAC=<base narrac> sh scripts/test.sh --suite lsp --filter hover_` fails on `hover_names` (template alias answers `string`, no hover on `Arrays`, `S`, `measure`, pattern `Shapes`, struct fields; `Point` and `Shape` instead of `struct Point` and the variant).
- branch: the same command passes (2 cases).
- Not reproduced as a fix: the missing `uses` line of a private function in a package with errors. The effect pass needs error-free bodies and does not run, in `narrac query hovers` too. `hover_errors` records this behavior and passes on base. Left, because inferring effects over bodies with errors is outside this node.
- Builtins: `hover_names` records `print`, `Int.toString` (signatures) and `Arrays.length` (no signature in the registry, so the call's type `int`).

## Deviations

- Two `narrac query type-of` lines in `testdata/query/basic/queries.txt` and one hover line in `testdata/query/hovers_names/queries.txt` changed, because they share `findTypeAt`/`collectHovers` (variant call text, module alias in a template, the `as` local name).

## Stubs left

- None.

## Gotchas

- `Arrays.length` style members with no registry signature show the call's type in the LSP and nothing in `query hovers`.
- A package with errors shows no inferred `uses` for private functions.

## Proposed learnings

- Put hover text in `describeSite` and the helpers beside it, not in `collectHovers` | the LSP and `query hovers` then agree | selfhost/src/tools | no

## Verification

- Suites lsp (24), query (16), sources (1) pass on the branch narrac.
- Full `sh scripts/test.sh`: see the final report.

## Size

- Estimated M, actual M.

## Review fixes

- A local named like a module alias is no longer shown as the module (`describeModuleUse`, `collectImportHovers`); case `lsp/hover_shadow`.
- A struct imported with `as` shows `struct Point as P`.
