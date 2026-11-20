# Wave 28 report: batch-20, pre-release fixes

Batch-19 (PR #124) merged and is set `done`. Batch-20 fixes what a first user meets before the first release tag, and reviews the docs. It runs on the local branch `feat/batch-20`, in the worktree `../narra-batch-20`.

## Landed

Each child landed as one commit after a worker check. M nodes also got a fresh review.

- narrac-version: `narrac --version` prints `dev` for a local build and the version for a release build (`scripts/release-build.sh` stamps a copy of `selfhost/`). `--help` and `-h` print the usage. The installers, `docs/install.md` and `release.yml` check with `--version`. The version has no leading `v`.
- const-float-exact: `@const` evaluation reads long float literals (`0.49999999999999994`, `1e300`, `5e-324`) through the C library reader. The `F32.from(int)` beyond 2^53 item was left: it needs a direct int-to-f32 conversion in the evaluator and the emitter.
- lsp-dependency-definition: go to definition into a git dependency returned null; `buildViews` now reads `SourceModule.source`. The member special case in `lspnavigation.nra` is gone.
- export-key-diagnostics: E0101, E0402 and E0421 and the `narrac pkg` dump name the export key the importer wrote. E0474 lists the exposed std paths. The remove-unused-import fix handles quotes inside `//` comments. The review was clean; two of its nits were applied.
- lsp-hover-parity: the LSP hover shows module aliases in template strings, struct literal names and fields, `enum Shape as S`, renamed import names, module aliases in patterns, and builtin signatures. The review found that a local named like a module alias hovered as the module; that is fixed, with the golden `hover_shadow`.
- regex-step-limits: `testdata/regex_linear` compares the time for n and 2n characters instead of a fixed second. Under 40 load processes the old case failed 10 of 12 runs and the new one 0 of 12. A backtracking matcher grows 876 times per doubling against the bound of 3.
- docs-structure-review: the proposal was approved by the maintainer on 2026-11-20, all changes.
- docs-release-pass: proposal changes 1 to 7 and an MIT `LICENSE`.

Dropped: effect-export-gap, not reproduced. The only shape that compiles is the entry function `run`, which is exempt from `uses` by design.

## Orchestrator rulings

- The C export chain (c-export-functions, c-abi-export, node-addon, wasm-js-glue) moved to `later` until a user asks for it (maintainer, 2026-11-19).
- first-release depends on batch-20.
- Nine reference and guide nodes from the docs proposal are outlines after first-release.
- A struct or enum under an alias hovers as `<kind> <Name> as <Alias>`.

## Checks

- `sh scripts/test.sh`: PASS, 10496 cases.
- `sh selfhost/bootstrap.sh`: fixed point, seed current (regenerated in `c8b6685f`).
- `pnpm --dir website build`: passes.
- `sh scripts/linux-check.sh`: PASS.

## Still to do

- [x] `sh scripts/linux-check.sh`: PASS in 1059s.
- [x] Push `feat/batch-20` and open the batch-20 pull request for the maintainer. PR #125, merged.
- [ ] After it merges: first-release. The maintainer decides the first tag (prerelease or full release, and the `install.sh` fallback for prereleases) and pushes it.
