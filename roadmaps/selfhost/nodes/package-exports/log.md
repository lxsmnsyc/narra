# Log: package-exports

- New `selfhost/src/exports.nra`: `rewriteImports` checks each import of another package against that package's `[exports]` table and replaces the key with the module path the value names (the AST copy only changes when an import is rewritten); `checkExportFiles` reports a value with no file (E0471); `isStdPathExposed` is the built-in std list.
- `manifest.nra`: `[exports]` is read into `Manifest.exports` (`ExportEntry`); key rules E0470 and E0472 and the value shape E0471 are checked at parse time. Keys may be quoted or bare.
- `depload.nra`: `DependencyPackage.exports` carries the table of each dependency.
- `modules.nra`: `loadPackage` builds the tables of the dependencies and runs `exposeImports` on every module of the package and of each dependency; one dispatch loop each.
- Why the loader rewrites the import path: scope lookup, module order and C names all key on the module path, so one rewrite before the checker keeps them unchanged. An import inside the same package is never rewritten.
- `scopes.nra`: the std check calls `isStdPathExposed` and reports E0474. E0426 is removed with no migration message. The message text stays `... is internal to the standard library`, so existing goldens did not change.
- `unused.nra`: the fix that removes an unused import looked for the line holding `"<path>"`; a rewritten import has another path in the AST, so it now stops at the first line with a quote.
- Codes E0470 to E0474 in `diagnostics.nra`.
- Fixtures: `testdata/deps/build/{exports,error-unexported,error-export-file}` (real file:// repositories) with cases `build-exports`, `build-error-unexported`, `build-error-export-file` in `depsbuild.nra`; `testdata/manifest/exports_*` for E0470, E0471 and E0472; `testdata/errors/exports_std_ext`.
- Decisions: duplicate key is E0472 (a TOML table with `"app"` twice is the "two keys for one bare name" case); an `[exports]` value that is not `src/...nra` is E0471 at parse time. Fixtures live next to the other dependency and manifest cases instead of `testdata/packages/exports_*`, because the runners are there.
