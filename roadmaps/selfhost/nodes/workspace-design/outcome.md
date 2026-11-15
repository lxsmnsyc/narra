# Outcome: workspace-design

## Built

- D101 in `DECISIONS.md`, between D100 and D102, marked "Not built yet". It decides to build workspaces after package-dependencies and package-exports, with a `[workspace] members` table, path dependencies between members, one `narra.lock` at the root, and path-selected commands. No compiler code.
- Five proposed child briefs below, for node workspaces. The last one moves this repository to a workspace (maintainer, 2026-11-14).

## Questions for the maintainer at the gate

1. **Root package.** D101 lets the root `narra.toml` hold `[package]` as well as `[workspace]` (Cargo). Keep that, or allow only a root with `[workspace]` alone (one fewer case to test)?
2. **Unlisted packages.** D101 treats a package under a workspace root that `members` do not list as standalone, so the 145 `testdata/` fixtures need nothing. Accept that, or require Cargo's error with an `exclude` key?
3. **Path dependency targets.** D101 allows a path dependency only to a member of the same workspace. Allow any directory with a `narra.toml`, also outside a workspace?
4. **Member against git dependency.** D101 makes a git dependency with a member's name an error. Should a member instead replace the git package, as `go.work` does?
5. **Git dependency on another repository's member.** Left undecided in D101. Needed before a library repository can be a workspace that others depend on.
6. **`query` at the root.** D101 makes it an error that names the members. Accept, or pick the root package when there is one?

## Proposed children

Codes E0480-E0489 are free (deps nodes hold E0430-E0479). Every child is priority could, milestone p4-packages, story none until the orchestrator places it.

### workspace-manifest

- **Title.** `[workspace]` table, path dependencies in `narra.toml`, and finding the workspace root.
- **Goal.** `parseManifest` reads `[workspace] members` and `json = { path = "../json" }`. A new module `selfhost/src/workspace.nra` exports `findWorkspace(packageDirectory): Option<Workspace>` and `listMembers(root): Result<Array<Member>, Array<LoadError>>` (each member's directory and name). Nothing loads members yet.
- **Size.** M.
- **Depends on.** deps-manifest-lock, kind contract, via dependencies (the inline-table reader and `Dependency`).
- **Owns.** `roadmaps/selfhost/nodes/workspace-manifest/**`, `selfhost/src/workspace.nra`, `testdata/workspaces/manifest_*/**`, `testdata/errors/workspace_*`.
- **Shared.** `selfhost/src/manifest.nra`, `selfhost/src/diagnostics.nra`, `selfhost/tests/src/manifest.nra`, `selfhost/tests/src/packages.nra`.
- **Reserves.** diag E0480-E0484. Contract change: `Dependency` gains `path: string` (empty for git), `contracts/dependencies.md` version 2, written as a change request for the orchestrator.
- **Acceptance.**
  - [ ] `members` takes exact paths and `*` as one whole segment. The members come out sorted by directory.
  - [ ] Errors in E0480-E0484, each with a case: `members` not an array of strings, a member with no `narra.toml`, two members with one name, a member with its own `[workspace]`, a path dependency with `git` or `version` beside `path`.
  - [ ] `findWorkspace` returns the first `[workspace]` above a package only when its `members` list the package. An unlisted package gets `None`.
  - [ ] A root with `[package]` and `[workspace]` lists itself as a member.

### workspace-loader

- **Title.** Load path dependencies, resolve from every member, one lock at the root.
- **Goal.** `Modules.load` on a member finds its workspace, loads each path dependency's `src/` as a dependency package (hashed with its name, checked against `[exports]`), runs deps-resolve from every member's manifest, and reads and writes `narra.lock` at the workspace root.
- **Size.** M.
- **Depends on.** workspace-manifest, kind hard (needs `findWorkspace` and `path`). package-dependencies, kind hard (the dependency loader and hash rule). deps-resolve, kind behavior, via dependencies (several roots). package-exports, kind behavior (exports between members are tested here).
- **Owns.** `roadmaps/selfhost/nodes/workspace-loader/**`, `testdata/workspaces/load_*/**`.
- **Shared.** `selfhost/src/modules.nra`, `selfhost/src/depresolve.nra`, `selfhost/src/hash.nra`, `selfhost/src/diagnostics.nra`, `selfhost/tests/src/packages.nra`, `testdata/deps/**`.
- **Reserves.** diag E0485-E0487.
- **Acceptance.**
  - [ ] A fixture workspace `app` -> `json` (path) -> `http` (git fixture) builds and runs under the sanitizers and leak check, from the `app` directory.
  - [ ] Two members that require different minimums of one git package get the higher one, whichever member is built. The root lock is byte-identical on a second run.
  - [ ] Errors in E0485-E0487, each with a case: a path dependency outside the workspace or naming a different package, a `narra.lock` inside a member, a fetched git package with a path dependency.
  - [ ] A member with `[exports]` exposes only its listed paths to another member.
  - [ ] Building a member alone emits the same C for its own modules as before the workspace existed (golden files untouched).

### workspace-commands

- **Title.** `build`, `run`, `test`, `check`, `fmt` and `query` at the workspace root.
- **Goal.** The driver commands act on every member when given the workspace root, on one member when given its directory, as in the D101 table. `fmt` uses the root `[fmt]` for members without one. SPEC section 15 and the guide's packages chapter describe workspaces.
- **Size.** M.
- **Depends on.** workspace-loader, kind hard (a member must load with its dependencies).
- **Owns.** `roadmaps/selfhost/nodes/workspace-commands/**`, `testdata/workspaces/commands_*/**`.
- **Shared.** `selfhost/src/driver.nra`, `selfhost/src/fmt.nra`, `selfhost/src/fmtconfig.nra`, `selfhost/src/query.nra`, `selfhost/src/main.nra`, `selfhost/src/diagnostics.nra`, `selfhost/tests/src/packages.nra`, `SPEC.md`, `docs/guide/*.md`.
- **Reserves.** diag E0488-E0489.
- **Acceptance.**
  - [ ] At the root: `build` makes every binary of every member, `test` runs each member's tests in `members` order, `check` reports every member's diagnostics sorted as today, `fmt --check` covers every member.
  - [ ] `run -bin NAME` at the root finds the binary across members. Errors: two members with one binary name (E0488), `query` at the root (E0489, the message lists the member directories).
  - [ ] `findConfig` skips a `narra.toml` with no `[package]` and no `[fmt]`, so files under `testdata/` keep the defaults once the repository root has a manifest.
  - [ ] SPEC section 15 documents `[workspace]`, path dependencies and the root lock. D101 drops "Not built yet" for the parts built.

### workspace-lsp

- **Title.** The language service checks a member and the members that depend on it.
- **Goal.** `narrac lsp` finds the package from the nearest manifest with `[package]`, finds its workspace, and after each change checks that member and its dependent members, each as its own program. Diagnostics go only to each program's own modules.
- **Size.** M.
- **Depends on.** workspace-loader, kind hard.
- **Owns.** `roadmaps/selfhost/nodes/workspace-lsp/**`, `testdata/lsp/workspace_*/**`.
- **Shared.** `selfhost/src/lspserver.nra`, `selfhost/src/lspcache.nra`, `selfhost/src/lspnavigation.nra`, `selfhost/tests/src/lsp.nra`.
- **Reserves.** none.
- **Acceptance.**
  - [ ] `findPackageDirectory` skips a manifest with no `[package]`. A single file under such a root is checked alone, as today.
  - [ ] An edit in member `json` publishes diagnostics for `json` and for member `app` that imports it. `app`'s publish holds no entry for `json`'s files.
  - [ ] An edit in `app` does not check `json`'s bodies again (the `lspcache` per-module state is keyed by package directory).
  - [ ] Go to definition from `app` into `json` opens the file in the member directory.

### repo-workspace

- **Title.** This repository becomes a workspace.
- **Goal.** Add a root `narra.toml` with `[workspace] members = ["selfhost", "selfhost/tests"]`, no `[package]` and no `[fmt]`. `narrac check .` and `narrac fmt --check .` at the root cover the compiler and the test runner. `std/`, `bench/` and `testdata/` stay outside.
- **Size.** S.
- **Depends on.** workspace-commands, kind hard. workspace-lsp, kind hard (without it, the lsp suite reaches the root manifest from `testdata/lsp`). selfhost-layout, kind soft (both edit `selfhost/CLAUDE.md`. Either order works).
- **Owns.** `roadmaps/selfhost/nodes/repo-workspace/**`, `narra.toml`.
- **Shared.** `selfhost/CLAUDE.md`, `README.md`, `scripts/test.sh`.
- **Reserves.** none.
- **Acceptance.**
  - [ ] `sh selfhost/bootstrap.sh` reaches the fixed point and reports the seed current with no seed change. `narrac emit selfhost` output is byte-identical before and after.
  - [ ] `sh scripts/test.sh` passes with the case total unchanged. The lsp, fmt and sources suites show no diff.
  - [ ] `narrac check .` at the root exits 0 and checks both members.
  - [ ] A test in the `packages` suite fails when the root manifest is removed or a member is dropped from it.

## Deviations

- None. No node folders were written for the children. They are in this file only.

## Stubs left

- None.

## Gotchas

- `moduleHash` with the package name prepended collides with the root package. Dependency `json`'s `src/parse.nra` would hash `"json/parse"`, the same input as the root package's `src/json/parse.nra`. The E0417 check in `loadPackage` (`selfhost/src/modules.nra` line 250) runs inside one package only. package-dependencies should hash a dependency module with a separator no module path holds (`json:parse`), or run the clash check over every package in the build.
- Adding any `narra.toml` at the repository root changes what `findPackageDirectory` (`selfhost/src/lspserver.nra` line 210) and `findConfig` (`selfhost/src/fmtconfig.nra` line 211) return for every file under `testdata/`. The repo move must not land before both skip manifests with no `[package]`.
- `contracts/dependencies.md` is frozen by deps-manifest-lock at version 1 with no `path` field. workspace-manifest needs version 2.
- `query` already has `--package DIR`. Do not add a driver flag with the same name and a different meaning.

## Proposed learnings

- A manifest walk up from a file must skip a `narra.toml` with no `[package]` | a workspace root manifest is not a package, and the repository root will have one | `selfhost/src` | yes, an lsp and a fmt case under a root with only `[workspace]`

## Verification

- No code changed, so `sh scripts/test.sh` was not run.
- `python3 <plain-writing>/scripts/check_prose.py` on D101 and this file: no errors.

## Size

- Estimated M, actual M.
