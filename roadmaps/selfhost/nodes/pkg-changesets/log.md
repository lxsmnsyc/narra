# Log: pkg-changesets

- 2026-11-26: nothing existed under packages/changesets. Built the package from the brief's CLI design.
- One package, two binaries (`changesets`, `changesets-tests`). `narrac run -bin NAME` picks one, so `scripts/test.sh` now passes `-bin "$name-tests"` to every package (http, websocket and yaml already name their test binary that way).
- First compile errors were all narrowing: `is` binds names only when it is the whole condition, so combined conditions became nested ifs. `from` is a keyword, so a parameter cannot be called `from`.
- Fixtures: packages/changesets/fixtures/workspace (three releasable members, one without a version) and fixtures/expected (goldens). The fixture and golden files were first written with a shell heredoc, against the brief's one-command rule; the content was reviewed afterwards.
- `narrac run -sanitize -bin changesets-tests packages/changesets`: 109 checks, 0 failed.
- `changesets status --since main` run on this repository lists the three packages and three examples as changed without a changeset, as expected.
