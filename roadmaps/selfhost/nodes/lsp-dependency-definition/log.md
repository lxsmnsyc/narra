# Log: lsp-dependency-definition

## Reproduction on the batch base

- Base: `feat/batch-20` at 94107ca4, narrac built with `sh selfhost/bootstrap.sh <dir>`.
- Fixture: a package `app` whose manifest has `json = { git = "file://<repo>", version = "1.0.0" }`, a local repository with one commit tagged `v1.0.0` holding `src/text.nra` with `export function render`, and `app/src/main.nra` calling `Text.render(...)` after `import * as Text from "json/text"`.
- `narrac build -o app.bin app` (fetches the dependency into `NARRA_CACHE`), then `narrac lsp` with an initialize, a didOpen of `main.nra`, and `textDocument/definition` on `render`.
- Result: `{"id":2,"result":null}`. `narrac query --package app def render` printed `error unknown name "render"`: the dependency's modules are missing from the views.
- Cause: `buildViews` read `<package directory>/<module.file>`; for a dependency the display file is `json/src/text.nra`, so the read failed and the module was skipped.

## Fix

- `SourceModule.source` (selfhost/src/driver/modules.nra) holds the path the text was read from. `loadPackageModules` sets it to `<directory>/src/<file>`, the single-file load to its path, std modules to their file name.
- `buildViews` reads and looks up overrides by `module.source`. It no longer skips unreadable dependency files.
- `lspnavigation.nra` takes each document's path from `module.source`; `mapMemberFiles` and the view overrides for members are gone, because member modules now carry their member directory path.

## Test

- Case `deps/build-definition` (testdata/deps/build/definition, selfhost/tests/src/depsbuild.nra): builds an app with a git dependency, checks `query def render` lists `json/text.render`, and checks an LSP definition request on `render` returns `src/text.nra` line 6, character 16 in the cache checkout.
- Base narrac: fails with `error unknown name "render"`. Branch narrac: passes.
