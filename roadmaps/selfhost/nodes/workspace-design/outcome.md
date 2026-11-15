# Outcome: workspace-design

## Built

- D101 in `DECISIONS.md`, between D100 and D102, marked "Not built yet". Revised on 2026-11-15 after the maintainer's answers and a study of pnpm. It decides:
  - `[workspace] members`, with `*` and `!` entries.
  - Member dependencies by name, `{ workspace = true }`.
  - No local directories outside the workspace.
  - Git dependencies on a member of another repository, tagged `<name>/v<version>`.
  - A root-only `[overrides]` table.
  - One `narra.lock` at the root, with commands selected by path plus `-dependents` and `-dependencies`.
- Proposed brief changes for deps-fetch and deps-resolve, and a contract change, below.
- Six proposed child briefs below, for node workspaces. The last one moves this repository to a workspace (maintainer, 2026-11-14).

## Maintainer answers (2026-11-15) and where D101 records them

| question | answer | in D101 |
| --- | --- | --- |
| 1. root package | yes | "Root package" |
| 2. unlisted packages standalone | yes | "Finding the workspace" |
| 3. path dependencies only to members | decide after pnpm | no `path` key at all, members by name only (see below) |
| 4. member with a git dependency's name is an error | yes | "Version selection" |
| 5. git dependency on another repository's member | yes | "Git dependency on a member of another repository" |
| 6. `query` at the root covers all members | yes | command table |

Question 3: pnpm's `link:` and `file:` take any directory (`pnpm11/resolving/local-resolver/src/parseBareSpecifier.ts` has no workspace check), and `pnpm publish` passes them through unchanged (`pnpm11/releasing/exportable-manifest/src/index.ts` rewrites only `workspace:` and `catalog:`). In Narra a fetched checkout is what other users get, so a directory outside the repository cannot work for them. D101 allows member dependencies only, by name.

## Second answers (2026-11-15)

The maintainer left the three follow-up questions to the orchestrator. D101 keeps its text on each:

1. **Member tags stay `<name>/v<version>`.** A dependency names a package, not a directory, and `listVersions(git, name)` looks tags up by that name. A tag by name survives a member moving to another directory, which the repository move to a workspace will do. Go tags by directory because a Go module path is a directory.
2. **Overrides may go below a minimum.** An override is written by hand at the root and recorded in the lock, so nothing is hidden. Its uses (moving off a broken release, pinning a fork) often need a lower version. Go's `replace` also ignores minimums. If the lower version lacks something a package needs, checking that package reports it.
3. **Flags stay `-dependents` and `-dependencies`.** narrac accepts both `-flag` and `--flag`, and `-filter` is taken by `narrac test`.

The gate is passed: children may start.

## Proposed contract change: `contracts/dependencies.md` version 2

- `Dependency` gains `workspace: bool`. When it is true, `git` and `version` are empty.
- `checkoutPath(git: string, tag: string)` replaces `checkoutPath(git, version)`. In a workspace repository, `json/v1.2.0` and `core/v1.2.0` are different commits, so a cache folder keyed by version alone would mix them. The folder becomes `<cache>/git/<sanitized url>/<sanitized tag>`.
- `LockEntry` is unchanged. The commit fixes the tag, and the package's directory is found again from the checkout's manifests.
- `[overrides]` entries reuse the `Dependency` shape.

## Proposed brief change: deps-fetch

Apply before deps-fetch starts. It is `todo` in batch-16.

- `fetchVersion(git, tag)` takes the tag, and the cache folder follows the contract change above.
- `listVersions(git, name)` returns the versions of tags `<name>/vMAJOR.MINOR.PATCH` when any exist, else those of tags `vMAJOR.MINOR.PATCH`.
- Acceptance adds a fixture repository with two members tagged `a/v1.0.0` and `b/v1.0.0` on different commits, fetched into two folders.

## Proposed brief change: deps-resolve

Apply before deps-resolve starts. It is `todo` in batch-16.

- **Several roots.** The resolver takes an array of root manifests, not one. For a standalone package the array has one entry. workspace-loader passes every member. This is a signature change only, with no workspace knowledge in deps-resolve.
- **`[overrides]`.** The root of the build may have `[overrides]`, with entries shaped like a `Dependency`. Every requirement on an overridden name uses the override's URL, and the selected version is the override's version exactly. An overridden name reached through two URLs is no longer an error, because the override replaces both URLs. The lock records the selected URL and version.
- **Codes.** Override errors take codes from E0450-E0459: `[overrides]` in a manifest that is not the root of the build, an override for a name no package requires, and an override whose tag is missing.
- **Acceptance adds.** A diamond fixture where the override pins the shared package to a fork URL and a lower version, and the lock shows the fork. An `[overrides]` table in a dependency's manifest is an error.
- **Manifest.** `parseManifest` reads `[overrides]` (owned with `manifest.nra` as a shared file). If deps-manifest-lock has already landed, deps-resolve adds the table.
- **If deps-resolve has started.** The same text becomes a child `deps-overrides`, size S, hard dependency on deps-resolve, owning `testdata/deps_overrides*`.

## Proposed children

Every child is priority could, milestone p4-packages, story none until the orchestrator places it. Codes E0480-E0489 are free (deps nodes hold E0430-E0479).

### workspace-manifest

- **Title.** `[workspace]` table, `workspace = true` dependencies, and finding the workspace root.
- **Goal.** `parseManifest` reads `[workspace] members` and `json = { workspace = true }`. A new module `selfhost/src/workspace.nra` exports two functions. `findWorkspace(packageDirectory): Option<Workspace>` finds the workspace. `listMembers(root): Result<Array<Member>, Array<LoadError>>` lists each member's directory and name. Nothing loads members yet.
- **Size.** M.
- **Depends on.** deps-manifest-lock, kind contract, via dependencies (version 2 above).
- **Owns.** `roadmaps/selfhost/nodes/workspace-manifest/**`, `selfhost/src/workspace.nra`, `testdata/workspaces/manifest_*/**`, `testdata/errors/workspace_*`.
- **Shared.** `selfhost/src/manifest.nra`, `selfhost/src/diagnostics.nra`, `selfhost/tests/src/manifest.nra`, `selfhost/tests/src/packages.nra`.
- **Reserves.** diag E0480-E0483.
- **Acceptance.**
  - [ ] `members` takes exact paths, `*` as one whole segment, and `!` entries that remove matches. Members come out sorted by directory.
  - [ ] Each of these errors has a case and a code in E0480-E0483:
    - `members` is not an array of strings.
    - A member has no `narra.toml`, or has its own `[workspace]`.
    - Two members have one name.
    - A `workspace = true` dependency has another key beside it.
  - [ ] `findWorkspace` returns the first `[workspace]` above a package only when its `members` list the package. An unlisted package gets `None`.
  - [ ] A root with `[package]` and `[workspace]` lists itself as a member.

### workspace-loader

- **Title.** Load member dependencies, resolve from every member, one lock at the root.
- **Goal.** `Modules.load` on a member finds its workspace and loads each `workspace = true` dependency's `src/` as a dependency package. Each such package is hashed with its name and checked against `[exports]`. The loader passes every member's manifest to deps-resolve and reads and writes `narra.lock` at the workspace root.
- **Size.** M.
- **Depends on.**
  - workspace-manifest, kind hard: it needs `findWorkspace` and `listMembers`.
  - package-dependencies, kind hard: the dependency loader and the hash rule.
  - deps-resolve, kind behavior, via dependencies: several roots.
  - package-exports, kind behavior: exports between members are tested here.
- **Owns.** `roadmaps/selfhost/nodes/workspace-loader/**`, `testdata/workspaces/load_*/**`.
- **Shared.** `selfhost/src/modules.nra`, `selfhost/src/diagnostics.nra`, `selfhost/tests/src/packages.nra`, `testdata/deps/**`.
- **Reserves.** diag E0484-E0485.
- **Acceptance.**
  - [ ] A fixture workspace builds and runs from the `app` directory under the sanitizers and leak check. In it, `app` depends on `json` by `workspace = true`, and `json` depends on the git fixture `http`.
  - [ ] Two members require different minimums of one git package. The higher one is selected, whichever member is built. The root lock is byte-identical on a second run.
  - [ ] Errors in E0484-E0485, each with a case. First, a `workspace = true` name with no such member, or a dependency outside a workspace. Second, a `narra.lock` inside a member. A git dependency with a member's name uses the deps-resolve error for one name with two sources.
  - [ ] A member with `[exports]` exposes only its listed paths to another member.
  - [ ] Building a member alone emits the same C for its own modules as before (golden files untouched).

### git-member-deps

- **Title.** Git dependencies on a member of another repository's workspace.
- **Goal.** A git dependency finds its package in the checkout: the root package, or else the member with that name. Tags are `<name>/v<version>` in a workspace repository. A `workspace = true` dependency inside a fetched checkout becomes a git requirement on the same URL, at the sibling's `[package] version`.
- **Size.** M.
- **Depends on.**
  - workspace-manifest, kind hard: it reads `[workspace]` in the checkout.
  - deps-fetch, kind behavior, via dependencies: the tag-keyed cache and `listVersions(git, name)`.
  - deps-resolve, kind behavior, via dependencies: it rewrites edges while walking.
  - package-dependencies, kind hard: it loads the found member's `src/`.
- **Owns.** `roadmaps/selfhost/nodes/git-member-deps/**`, `testdata/deps/workspace_*/**`.
- **Shared.** `selfhost/src/depresolve.nra`, `selfhost/src/depcache.nra`, `selfhost/src/modules.nra`, `selfhost/src/diagnostics.nra`, `selfhost/tests/src/gitfixtures.nra`, `SPEC.md`.
- **Reserves.** diag E0486-E0487.
- **Acceptance.**
  - [ ] A fixture repository with members `json` and `core` is tagged `json/v1.1.0` and `core/v1.0.0`, and `json` has `core = { workspace = true }`. An app with `json = { git, version = "1.1.0" }` builds with `core` 1.0.0 in its lock.
  - [ ] The same app also requiring `core` 1.2.0 selects 1.2.0.
  - [ ] Errors in E0486-E0487, each with a case. First, no package of that name in the checkout. Second, a sibling with no `version`, or no tag for that version.
  - [ ] SPEC section 15 documents member tags.

### workspace-commands

- **Title.** `build`, `run`, `test`, `check`, `fmt` and `query` at the workspace root.
- **Goal.** The driver commands act on every member when given the workspace root, and on one member when given its directory, as in the D101 table. `-dependents` and `-dependencies` add related members. `fmt` uses the root `[fmt]` for members without one. SPEC section 15 and the guide's packages chapter describe workspaces.
- **Size.** M.
- **Depends on.** workspace-loader, kind hard.
- **Owns.** `roadmaps/selfhost/nodes/workspace-commands/**`, `testdata/workspaces/commands_*/**`.
- **Shared.** `selfhost/src/driver.nra`, `selfhost/src/fmt.nra`, `selfhost/src/fmtconfig.nra`, `selfhost/src/query.nra`, `selfhost/src/main.nra`, `selfhost/src/diagnostics.nra`, `selfhost/tests/src/packages.nra`, `SPEC.md`, `docs/guide/*.md`.
- **Reserves.** diag E0488-E0489.
- **Acceptance.**
  - [ ] At the root, each command covers every member:
    - `build` makes every binary of every member.
    - `test` runs each member's tests in `members` order.
    - `check` reports every member's diagnostics, sorted as today.
    - `fmt --check` covers every member.
    - `query def NAME` searches every member.
  - [ ] `test packages/json -dependents` also runs the tests of `app`. `check packages/app -dependencies` also checks `json`.
  - [ ] `run -bin NAME` at the root finds the binary across members. Two members with one binary name is E0488. A flag `-dependents` or `-dependencies` with a path that is not a member is E0489.
  - [ ] `findConfig` skips a `narra.toml` with no `[package]` and no `[fmt]`, so files under `testdata/` keep the defaults.
  - [ ] SPEC section 15 documents `[workspace]`, `workspace = true`, `[overrides]` and the root lock. D101 drops "Not built yet" for the parts built.

### workspace-lsp

- **Title.** The language service checks a member and the members that depend on it.
- **Goal.** `narrac lsp` finds the package from the nearest manifest with `[package]`, then finds its workspace. After each change, it checks that member and its dependent members, each as its own program. Diagnostics go only to each program's own modules.
- **Size.** M.
- **Depends on.** workspace-loader, kind hard.
- **Owns.** `roadmaps/selfhost/nodes/workspace-lsp/**`, `testdata/lsp/workspace_*/**`.
- **Shared.** `selfhost/src/lspserver.nra`, `selfhost/src/lspcache.nra`, `selfhost/src/lspnavigation.nra`, `selfhost/tests/src/lsp.nra`.
- **Reserves.** none.
- **Acceptance.**
  - [ ] `findPackageDirectory` skips a manifest with no `[package]`. A single file under such a root is checked alone, as today.
  - [ ] An edit in member `json` publishes diagnostics for `json` and for member `app` that imports it. `app`'s publish holds no entry for `json`'s files.
  - [ ] An edit in `app` does not check `json`'s bodies again. The `lspcache` per-module state is keyed by package directory.
  - [ ] Go to definition from `app` into `json` opens the file in the member directory.

### repo-workspace

- **Title.** This repository becomes a workspace.
- **Goal.** Add a root `narra.toml` with `[workspace] members = ["selfhost", "selfhost/tests"]`, no `[package]` and no `[fmt]`. `narrac check .` and `narrac fmt --check .` at the root cover the compiler and the test runner. `std/`, `bench/` and `testdata/` stay outside.
- **Size.** S.
- **Depends on.**
  - workspace-commands, kind hard.
  - workspace-lsp, kind hard. Without it, the lsp suite reaches the root manifest from `testdata/lsp`.
  - selfhost-layout, kind soft. Both edit `selfhost/CLAUDE.md`, and either order works.
- **Owns.** `roadmaps/selfhost/nodes/repo-workspace/**`, `narra.toml`.
- **Shared.** `selfhost/CLAUDE.md`, `README.md`, `scripts/test.sh`.
- **Reserves.** none.
- **Acceptance.**
  - [ ] `sh selfhost/bootstrap.sh` reaches the fixed point and reports the seed current with no seed change. `narrac emit selfhost` output is byte-identical before and after.
  - [ ] `sh scripts/test.sh` passes with the case total unchanged. The lsp, fmt and sources suites show no diff.
  - [ ] `narrac check .` at the root exits 0 and checks both members.
  - [ ] A test in the `packages` suite fails when the root manifest is removed or a member is dropped from it.

## Deviations

- The first commit's subject was shortened to `workspace-design: Spike: workspaces for several packages in one repo`. The plain-writing hook rejects subjects over 72 characters, and the requested subject has 74.
- No node folders were written for the children. They are in this file only.
- No edits to the deps-fetch or deps-resolve briefs or the contract. The changes are proposals above.

## Stubs left

- None.

## Gotchas

- `moduleHash` with the package name prepended collides with the root package. Dependency `json`'s `src/parse.nra` would hash `"json/parse"`. That is the same input as the root package's `src/json/parse.nra`. The E0417 check in `loadPackage` (`selfhost/src/modules.nra` line 250) runs inside one package only. package-dependencies should hash a dependency module with a separator no module path holds (`json:parse`), or run the clash check over every package in the build.
- Adding any `narra.toml` at the repository root changes what two lookups return for every file under `testdata/`. They are `findPackageDirectory` (`selfhost/src/lspserver.nra` line 210) and `findConfig` (`selfhost/src/fmtconfig.nra` line 211). The repo move must not land before both skip manifests with no `[package]`.
- `contracts/dependencies.md` is frozen by deps-manifest-lock at version 1, keyed by version alone. Members of one repository need the tag-keyed version 2 before deps-fetch writes its cache.
- `query` already has `--package DIR`, and `narrac test` already has `-filter` for test names. The new selection flags avoid both names.

## Proposed learnings

- A manifest walk up from a file must skip a `narra.toml` with no `[package]` | a workspace root manifest is not a package, and the repository root will have one | `selfhost/src` | yes, an lsp case and a fmt case under a root with only `[workspace]`

## Verification

- No code changed, so `sh scripts/test.sh` was not run.
- `check_prose.py` (plain-writing) on D101 and this file: no errors.

## Size

- Estimated M, actual M.
