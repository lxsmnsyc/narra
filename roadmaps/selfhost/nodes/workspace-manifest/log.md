# Log: workspace-manifest

- Branch feat/workspace-manifest reset to feat/batch-18. Nothing of the work existed.
- manifest.nra: `[workspace] members`, `Dependency.workspace`, `Manifest.isWorkspace` and `members`. A manifest with `[workspace]` needs no `[package]`. Inline tables take bare `true` and `false`.
- workspace.nra: `findWorkspace`, `listMembers`, and the test-only command `narrac pkg --codes --workspace <dir>`.
- Tests: testdata/workspaces/manifest_* run through the manifest suite (variant workspace; a `start` file names the directory to start from).
- Review fixes: membership is decided from the member paths alone (no member manifests read) and from the climbed path, with the working directory from `pwd -P`; a `[package]` without a name is E0414 even with `[workspace]`; empty and `.` member entries are E0480; nine more cases (the `cwd` file runs the command inside a folder).
