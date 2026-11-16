# Outcome: workspace-loader

## Built

- `selfhost/src/driver/depload.nra`: `prepareDependencies` now calls `findWorkspace`. In a workspace it lists the members, reads every member manifest, resolves git requirements from all of them at once (the root's `[overrides]` apply), and reads and writes `narra.lock` at the workspace root. It returns the members the built package reaches through `workspace = true` (directly or not), each as a `DependencyPackage` with its `[exports]`, followed by the git packages those members and the built package reach. Outside a workspace nothing changes. New helpers: `prepareGitPackages` (the old body, now taking roots and overrides), `prepareWorkspaceDependencies`, `checkMembers`, `pickReachablePackages`, `keepGitDependencies`, `findMemberIndex`, `readMemberManifest`.
- No change in `modules.nra`: it already hashes every non-root package with its name (`json:parse`), runs the exports check and the undeclared-import check for each `DependencyPackage`. Members are `DependencyPackage` values, so they go through the same path.
- Codes: E0484 (a `workspace = true` name with no such member, or `workspace = true` outside a workspace), E0485 (a `narra.lock` inside a member). Rows in `selfhost/src/driver/diagnostics.nra`, plus E0450 for a git dependency, direct or deeper in the git graph, that has a member's name, and E0454 for `[overrides]` in a member other than the root (both reuse existing codes).
- Cases: `testdata/workspaces/load_*` (9 fixtures), run by the `packages` suite as `packages/workspaces/load_*` (`selfhost/tests/src/packages.nra`, function `runLoadCase`). A fixture holds `root/` (the workspace), optional `repos/` (git fixtures), `run.out` or `error.txt`, optional `runs` (members to run) and `lock.has`.

## Acceptance

- Fixture workspace from the `app` directory: `load_basic`. `app` depends on `json` by `workspace = true`, `json` depends on the git package `http`. It runs with `-sanitize` and with `-release`. `app` also has `src/json/parse.nra`, which does not clash with the member's `json/parse` (hash `json:parse`).
- Two members, two minimums: `load_versions`. `a` asks `http` 1.0.0, `b` asks 1.1.0. Both builds select 1.1.0; the root lock after building `a`, and after removing it and building `b`, is byte-identical, and a second build leaves it unchanged.
- Errors: `load_error_no_member` and `load_error_outside` (E0484), `load_error_member_lock` (E0485), `load_error_git_member_name` (E0450, the deps-resolve code for one name from two sources; `load_error_git_member_transitive` for a package deeper in the graph).
- Exports between members: `load_exports` imports the listed path; `load_error_unexported` imports `lib/internal` and gets E0473.
- A package built alone emits the same C: no golden changed; the full run passes.

## Deviations

- Cases are in the `packages` suite (`selfhost/tests/src/packages.nra`, named in the brief's shared paths), not in the `deps` suite.
- E0454 is also reported for `[overrides]` in a non-root member (an existing code, no new number).
- The test output has no codes, so each error case matches a part of the message (`error.txt`).
- `modules.nra` and `testdata/deps/` are untouched.

## Stubs left

- A `workspace = true` dependency inside a fetched git checkout (a member of another repository): node git-member-deps.
- Commands at the workspace root (`build`, `run`, `test` over every member): node workspace-commands. A root without `[package]` is not a buildable package here.

## Gotchas

- All members are validated on every load, so a broken sibling member (bad manifest, `narra.lock` inside it, unknown `workspace = true` name) fails the build of a healthy member.
- A git package is loaded only when the built package or a member it reaches depends on it. Packages in the root lock that only other members use stay out of the program, so their unused names are not errors.
- A cycle of `workspace = true` dependencies between members (a uses b, b uses a) is not detected here. It shows up later as an undeclared-import error. workspace-commands should report it.
- With no git requirement anywhere in the workspace, no lock is written.
- The language service (`fetch` false) checks that the lock covers every member's git requirements.

## Proposed learnings

- None.

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite packages --filter workspaces`: `test.sh: PASS, 8 cases`.
- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C.
- `sh scripts/test.sh`: `test.sh: PASS, 10279 cases in 236s`.

## Size

- Estimated M, actual M.
