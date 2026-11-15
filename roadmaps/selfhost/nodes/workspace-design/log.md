# Log: workspace-design

- 2026-11-15: started on `spike/workspace-design`, reset to `feat/batch-16` 9f4017d7.
- Read `manifest.nra`, `modules.nra` (`loadPackage`), `hash.nra` (`moduleHash`), `driver.nra` (`runBuild`), `lspserver.nra` (`findPackageDirectory`, `checkDocumentPackage`), `fmtconfig.nra` (`findConfig`), `query.nra` (`--package DIR`) and SPEC section 15.
- Repo: `selfhost/narra.toml` (narrac) and `selfhost/tests/narra.toml` (testrunner) are the only packages outside `testdata/`. `testdata/` holds 145 `narra.toml` fixtures. `std/` has no manifest. Its sources are embedded through `stdtext.nra`. `bench/` holds single files next to `.go` and `.rs` twins.
- Finding: once the repo root has a `narra.toml`, `findPackageDirectory` and `findConfig` reach it from every file under `testdata/`. A manifest with no `[package]` must be skipped by both, or the lsp and fmt suites change.
- Finding: `moduleHash` with the package name prepended (`json/parse`) equals the root package's hash for `src/json/parse.nra`. The E0417 check in `loadPackage` runs inside one package only. Recorded as a gotcha for package-dependencies.
- Finding: the testrunner does not import any `narrac/` module, so the repo move needs no member dependency.
- Wrote D101 between D100 and D102 in `DECISIONS.md`, and `outcome.md` with five proposed children.
- 2026-11-15: maintainer answers: 1, 2, 4, 5 and 6 yes, 3 after a pnpm study. Read pnpm docs (workspaces, pnpm-workspace_yaml, filtering, catalogs, settings/dependency-resolution, cli/link) and source (`pnpm11/releasing/exportable-manifest/src/index.ts`, `pnpm11/resolving/local-resolver/src/parseBareSpecifier.ts`).
- pnpm source: `link:` and `file:` have no workspace check and pass through publish unchanged. `replaceWorkspaceProtocol` turns `workspace:*` into the exact local version at publish.
- Revised D101: members by name (`workspace = true`, no `path`), `!` in `members`, member tags `<name>/v<version>`, fetched `workspace = true` turned into a git requirement, root-only `[overrides]`, `-dependents` and `-dependencies`. Catalogs, per-edge overrides, removal overrides, injected deps and git-ref filters rejected.
- Outcome: contract version 2, brief changes for deps-fetch and deps-resolve, six children (new: git-member-deps).
