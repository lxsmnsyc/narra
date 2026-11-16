# batch-19 log

## Fixes from the whole-branch review

Before is the compiler built from feat/batch-19, after is the compiler built from this branch. Each case is a step in `testdata/workspaces/*/steps` (or `error.txt`), run by the `packages` suite.

| Finding | Case | Before | After |
| --- | --- | --- | --- |
| 4 | `commands_query_errors` | `query --package . def f` prints `error src/main.nra:2:10: ...`; `--json` has `"file": "src/main.nra"` | `error c/src/main.nra:2:10: ...`; `"file": "c/src/main.nra"` |
| 5 | `commands_error_flags` | `one/src/main.nra/narra.toml:0:1: -dependents needs ...`, `loose/narra.toml:0:1: ...` | `one/src/main.nra:0:1: ...`, `loose:0:1: ...` |
| 6 | `load_error_git_member_transitive` | `required from <cache checkout directory>` | `required from file://<repository>` |
| 9 | `commands_error_fmt` | `error narra.toml: the member "missing" has no narra.toml` | `error narra.toml:0:1: the member "missing" has no narra.toml` |
| 10 | `commands_error_binary` | `run .` prints `the workspace has 2 binaries; choose one with -bin` | `run .` prints E0488, `the members one and two both have a binary named same` |

Finding 11 is not rebuilt from the views. Each query command writes its JSON as text inside the member's own run, so a merge from structured data means changing six commands to return values. SPEC now says what the code does: the lists of every member are merged, and `name` and `module` are the first answering member's. `type-of` names a file, so one member answers.

Finding 8: `readManifestAt`, `isRootWithoutPackage` and `findWorkspaceRoot` are exported from `driver/workspace.nra` and used by `depload`, `workspaceselect`, `lspserver` and `runWorkspaceCommand`. `fmtconfig` is left alone. It can import the driver group, but `parseManifest` rejects a manifest that has `[fmt]` and no `[package]` or `[workspace]` (E0414), and `fmt` must accept that file; `Manifest` also does not record `[fmt]`. A first try with the shared reader failed `packages/workspaces/commands_fmt` step 4. `findMemberIndex` stays: it returns an index into the member list, which `hasName` does not. `containsName` in `lspserver` is replaced by `hasName`.

Finding 7: `listDependentPrograms` calls `selectMembers(directory, true, false, false)`; an `Err` gives the member alone. The lsp workspace cases pass.
