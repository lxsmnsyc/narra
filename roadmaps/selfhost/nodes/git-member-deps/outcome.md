# Outcome: git-member-deps

## Built

- `selfhost/src/driver/depresolve.nra`:
  - `findCheckoutPackage(checkout, name, version)`: the root package of a checkout, or else the `[workspace]` member with that name. Returns `CheckoutPackage { directory, manifest, members }`. E0486 when a workspace root has no such package; a root that is not a workspace and names another package is still E0453.
  - A `workspace = true` dependency in a fetched manifest becomes a `Requirement` on the same URL at the sibling's `[package] version`. E0487 when the sibling has no version, or has no tag for it (E0451 of the lookup is turned into E0487 for these requirements).
  - Requirements of the build's own manifest skip `workspace = true` entries (they are not fetched; the workspace loader handles them).
  - The cache check takes, without git: the tag of a lock entry with the same name, URL and version; a cached member tag `<name>/v<version>`; and a cached plain tag `v<version>` when the lock has an entry of the same name and URL whose tag is plain. Otherwise it lists the tags with `git ls-remote` and picks the member tag first, then the plain tag.
- `selfhost/src/driver/depload.nra`: `listPackages` finds each dependency through `findCheckoutPackage`, so `DependencyPackage.directory` is the member folder and its `src/` loads.
- `selfhost/src/driver/diagnostics.nra`: E0486, E0487 (two messages).
- `SPEC.md` section 15: paragraph "A git dependency on a member of another repository".
- Cases in the `deps` suite (`selfhost/tests/src/depsresolve.nra`, `depsbuild.nra`; fixtures `testdata/deps/workspace_member_*`):
  - `resolve-member-basic`: `json/v1.1.0` and `core/v1.0.0` tags, `json` has `core = { workspace = true }`; the lock has core 1.0.0.
  - `resolve-member-higher`: the app also needs `core` 1.2.0; the lock has core 1.2.0.
  - `resolve-member-nopackage` (E0486), `resolve-member-nosibling` (E0487, no version), `resolve-member-notag` (E0487, no tag).
  - `resolve-member-cache`: the cache tag fix. `resolve-member-offline`: a locked rebuild with PATH empty over plain tags and an unselected version. `resolve-member-sibling`: `workspace = true` naming a sibling the checkout lacks (E0486).
  - `build-member`: builds and runs an app whose `json` member imports `core/base`, with the lock naming both member tags.

## Proof

- base: commit 8a5d7abc (feat/batch-19), `NARRAC=out/narrac sh scripts/test.sh --suite deps --filter resolve-member-cache` with only the new case: FAIL. The repository has tags `v1.2.0` and `json/v1.2.0`; with `v1.2.0` already cached, the lock records `tag = "v1.2.0"`, while a clean machine records `json/v1.2.0`.
- branch: the same command: PASS (lock has `tag = "json/v1.2.0"`, commit and tree equal the member tag's). `--suite deps`: `PASS, 52 cases`.

## Deviations

- Touched outside the owned and shared lists: `selfhost/src/driver/depload.nra` is not in the brief's shared paths (the member folder must reach the loader); `selfhost/tests/src/depsresolve.nra` and `depsbuild.nra` hold the cases.
- `build-member` takes its fixture from `testdata/deps/workspace_member_build` through the path `../workspace_member_build`, because the owned fixture paths are `testdata/deps/workspace_*`.
- Git runs for a version only when no lock entry (same name and URL) and no cached member tag covers it, and the tag is not cached: a cached plain tag with no plain lock entry for that name and URL costs one `git ls-remote`, so the tag is chosen as a clean machine chooses it. A locked rebuild runs no git, including for versions the walk reads but the lock does not select (case `resolve-member-offline`).
- No change to `modules.nra`, so the workspace-loader rebase should be clean.

## Stubs left

- `checkLockCovers` in `depload.nra` (language service) still reads the build manifest's `workspace = true` entries as git requirements with an empty URL. The workspace loader handles member dependencies of the build's own workspace.

## Gotchas

- A member of a fetched workspace pins its siblings at the versions in its own commit: tag the sibling before the dependent.
- The lock stays sorted by name, so a case's `expected.txt` lists packages alphabetically.
- Seed not regenerated.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh --suite deps`: PASS, 52 cases.
- `sh scripts/test.sh`: `test.sh: PASS, 10345 cases in 388s` (after the rebase onto workspace-loader).

## Size

- Estimated M, actual M.

Landed on feat/batch-19 as abc8b76b.
