# Outcome: lsp-navigation

## Built

- `selfhost/src/lspnavigation.nra`: `isNavigationRequest(method)` and `answerNavigation(server, method, params): Result<Value, string>` for `textDocument/hover`, `definition`, `references` and `documentSymbol`. An Err becomes -32602 in `lsp.nra`.
- `selfhost/src/lsp.nra`: one `if` in `handleRequest`, four `*Provider: true` capabilities in `handleInitialize`, one import.
- `selfhost/src/query.nra`: lookups that return data, used by both `narrac query` and the server.
  - `ModuleView` (now with `text` and `tokens`), `Query`, `Declared`, `Target`, `Use` exported.
  - `buildViews(checker, package, overrides)`: reads a module from `overrides` (key `DIR/src/x.nra`, as `Modules.loadWith`) before disk. `runQuery` passes `Map {}`.
  - `findTypeAt(mutable checker, mutable effects, effectsKnown, view, cursor): TypeAt` (`Found(TypeAnswer { location, typeText, effects })`, `InTemplate(start)`, `Nothing`). `queryTypeOf` formats it.
  - `findUses(checker, query, target): Array<Use>`. `queryRefs` formats it.
  - `findNodeBody`, `siteLocation`, `listDeclarations`, `findTarget` exported unchanged.
  - `narrac query` output is byte-identical (query suite, no golden change).
- `selfhost/src/lspserver.nra`: `countUnits`, `findPackageDirectory`, `makePosition` exported (the word `export` only).
- Tests (`lsp` suite): `testdata/lsp/nav_hover`, `nav_definition`, `nav_references`, `nav_symbols`, `nav_stale`. Each is the same two-module package (`src/main.nra`, `src/text.nra`). The existing lsp `expected.txt` files changed only in the initialize capabilities.
- Docs: compiler/README.md "Language server".

## Behavior to know (for lsp-incremental)

- Each navigation request runs `loadNavigation`: it builds overrides from `server.documents`, then `Modules.loadWith`, `Body.checkPackage`, `Effects.checkEffects` (only when the bodies have no errors) and `buildViews`. That is a full check per request, on top of the one per change. The obvious cache is the `Navigation` of the last check, kept until the next open, change or close. lsp-incremental can keep it on the `Server`; I did not add a field there, so that `Server` stays as it is for that node.
- When the check has errors (an unused parameter is one), answers still come from the checker tables; `effectsKnown` is false and hover leaves out the effects of expressions (a call's `uses` line). The type of a named private function then has no `uses`, because its inferred effects are unknown.
- Positions: LSP to compiler with `countCharacters` (the inverse of `countUnits`) on the document line; compiler to LSP with `countUnits`. Lines come from `ModuleView.text`, so the unsaved text sets them.
- Result URIs: the open document's URI when the file is open, else `convertPathToUri`.

## Acceptance

- [x] Hover: `nav_hover` ids 2 to 12. A function name (`(string) => int uses alloc`), the `(` of a call (`int`, `uses alloc`), locals, a field read, a constant, a parameter, a pattern binding; null on a blank line and on a keyword.
- [x] Definition: `nav_definition`. A function (cross-module, text.nra from disk), struct (annotation and literal), constant, variant (pattern and call), enum (parameter type), distinct type (`Text.UserId`), a name in an import list, a trait (implement block), a parameter, a pattern binding and a `const`. Null for a std function, a name inside a template and a keyword.
- [x] References: `nav_references`. A function with and without `includeDeclaration`, from a use and from the declaration, uses in both modules and inside a template string; a constant; a constructor; `[]` for a struct, a local and a blank line.
- [x] Document symbols: `nav_symbols`, for the closed text.nra (from disk) and main.nra. Kinds, full ranges, fields and variants as children.
- [x] Encoding and unsaved edits: `nav_hover` uses UTF-16 after `"é😀"` (`point` at 32, LIMIT at 42). `nav_stale` negotiates UTF-8 (`point` at 35), changes main.nra (a line inserted) and text.nra (lines inserted), and the answers follow both texts.
- [x] Template string: hover inside `${...}` gives the whole template (`nav_hover` id 10, range 8 to 38, `string`, `uses alloc`). References report a use there with the whole template's range.
- [x] Fast check: lsp, query, sources, programs, errors, bootstrap pass (see the report).

## Gate decisions

- Hover inside an interpolated template answers for the whole template (its type and range). Alternatives: `null` (what `query` does: an error), or the innermost node (it has no position of its own). Chosen because the brief asks for the whole template, and an editor shows something useful.
- Navigation answers a package whose bodies have errors, without the effect pass. Alternative: `null` as `query` exits 1. Rejected: an unused parameter while typing would turn every hover off.
- Definition of a type, trait or imported name with no site under the cursor uses the identifier token and the module scope. Alternative: add type-expression sites to `querysites.nra`. Rejected for now: larger change, and `query` would need it for `refs` of types anyway (left as is, see query-cli limits).
- Symbol ranges run to the last token before the next declaration. Alternative: name range only. A full range lets the editor's outline follow the cursor.
- Distinct types are SymbolKind Class (5). LSP has no distinct-type kind; TypeParameter (26) would be misleading.
- Recheck per request, no cache (see above). Alternative: a cache on `Server` now. Left for lsp-incremental, which owns the recheck.

## New codes

- None.

## Stubs left

- No definition for implement block functions, trait functions or std declarations; no references for types, traits, locals or parameters (query-cli limit).
- Template strings: no definition inside `${...}`; hover gives the whole template. Both follow from the parser giving nodes inside a template the template's position (inbox item from query-cli).

## Size

- Estimated M, actual M (about 640 lines of Narra, half in lspnavigation.nra).
