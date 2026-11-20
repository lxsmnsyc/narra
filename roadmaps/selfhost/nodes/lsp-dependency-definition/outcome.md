# Outcome: lsp-dependency-definition

## Built

- `SourceModule.source` in `selfhost/src/driver/modules.nra`: the path a module's text was read from.
- `buildViews` in `selfhost/src/tools/query.nra` reads and looks up overrides by `module.source`, so dependency modules (git cache, path, member) get views. `narrac query def` now lists their declarations.
- `selfhost/src/tools/lspnavigation.nra`: document paths come from `module.source`; `mapMemberFiles` and the member view overrides are removed.
- Case `deps/build-definition` (`testdata/deps/build/definition`, `selfhost/tests/src/depsbuild.nra`).

## Proof

- base: 94107ca4 (`feat/batch-20`), `NARRAC=<base narrac> sh scripts/test.sh --suite deps --filter build-definition` fails with `error unknown name "render"`. By hand, an LSP definition request on `Text.render` into a git dependency returns `null`.
- branch: the same command passes (1 case). The LSP request returns the cache checkout's `src/text.nra`, line 6, character 16.

## Deviations

- Added a field to `SourceModule` and a parameter to `newModule` in `selfhost/src/driver/modules.nra`, which is outside the shared paths. It is the minimal way to carry the real path to `buildViews`.
- The test case lives in the deps suite (it needs the git fixtures), not under `testdata/lsp`.

## Stubs left

- None.

## Gotchas

- `ModuleView.file` stays the display path (`json/src/text.nra`); use `module.source` for the real path.
- An unreadable dependency file now makes `buildViews` return an error instead of skipping the module.

## Proposed learnings

- Keep the real source path on the module, do not rebuild it from the display path | the display path of a dependency is not under the package directory | selfhost/src | no

## Verification

- Suites lsp (22), query (16), sources (1), deps (55) pass on the branch narrac.
- Full `sh scripts/test.sh`: see the final report.

## Size

- Estimated S, actual S.
