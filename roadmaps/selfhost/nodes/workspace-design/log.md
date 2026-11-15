# Log: workspace-design

- 2026-11-15: started on `spike/workspace-design`, reset to `feat/batch-16` 9f4017d7.
- Read `manifest.nra`, `modules.nra` (`loadPackage`), `hash.nra` (`moduleHash`), `driver.nra` (`runBuild`), `lspserver.nra` (`findPackageDirectory`, `checkDocumentPackage`), `fmtconfig.nra` (`findConfig`), `query.nra` (`--package DIR`) and SPEC section 15.
- Repo: `selfhost/narra.toml` (narrac) and `selfhost/tests/narra.toml` (testrunner) are the only packages outside `testdata/`. `testdata/` holds 145 `narra.toml` fixtures. `std/` has no manifest. Its sources are embedded through `stdtext.nra`. `bench/` holds single files next to `.go` and `.rs` twins.
- Finding: once the repo root has a `narra.toml`, `findPackageDirectory` and `findConfig` reach it from every file under `testdata/`. A manifest with no `[package]` must be skipped by both, or the lsp and fmt suites change.
- Finding: `moduleHash` with the package name prepended (`json/parse`) equals the root package's hash for `src/json/parse.nra`. The E0417 check in `loadPackage` runs inside one package only. Recorded as a gotcha for package-dependencies.
- Finding: the testrunner does not import any `narrac/` module, so the repo move needs no member dependency.
- Wrote D101 between D100 and D102 in `DECISIONS.md`, and `outcome.md` with five proposed children.
