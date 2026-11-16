# Outcome: workspace-manifest

## Built

- `selfhost/src/driver/workspace.nra`: `findWorkspace(packageDirectory): Option<Workspace>`, `listMembers(workspace): Result<Array<Member>, Array<LoadError>>`, structs `Workspace { root, manifest }` and `Member { directory, name }`, and the test-only command `narrac pkg --codes --workspace <dir>`.
- `selfhost/src/driver/manifest.nra`: `[workspace] members`, `{ workspace = true }` dependencies (`Dependency.workspace`), `Manifest.isWorkspace` and `Manifest.members`.
- Codes E0480 (members not an array of strings, unknown key, missing members, bad entry), E0481 (member without narra.toml, member with its own `[workspace]`), E0482 (two members with one name), E0483 (`workspace = true` with another key, `workspace` not true, `workspace` in an override).
- Cases: `testdata/workspaces/manifest_*` (22), run by the manifest suite.

## Deviations

- `selfhost/src/main.nra` edited: one import and one dispatch line for the test-only command. It is not in the brief's paths.
- `testdata/workspaces/manifest_*` holds the cases instead of `testdata/errors/workspace_*`, because the errors suite takes single `.nra` samples.
- `listMembers` takes the `Workspace` that `findWorkspace` returns, not a bare path.
- `findWorkspace` returns `Some` when the root's members cannot be listed, so the caller's `listMembers` reports the error.
- Exclusions (`!`) apply to every included entry regardless of order.

## Stubs left

- Nothing reads `Dependency.workspace` yet: deps-resolve and the loader see an empty `git` for such a dependency. The later workspace nodes handle it.

## Gotchas

- `findWorkspace` is lexical: it climbs path names and never resolves symlinks, so a package reached through a symlink is matched by the path it was given. A relative start is made absolute with the shell's `pwd -P`, not `PWD`.
- A `*` segment matches only child directories that have a narra.toml.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: PASS (see the final report).

## Size

- Estimated M, actual M.
