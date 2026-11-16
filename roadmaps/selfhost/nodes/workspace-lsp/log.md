# Log: workspace-lsp

- lspcache: per-module state keyed by package directory (paths, valid, results keys).
- lspserver: findPackageDirectory skips manifests with no [package]; checkDocumentPackage checks the member, then each dependent member, each through checkProgram.
- fixtures: testdata/lsp/workspace_members, workspace_single_root.
- review fixes: before the fix, a syntax error in json (`return (`) was also published on `file://$ROOT/app/json/src/parse.nra` (workspace_members expected.txt held that line, 1 match for app/json). checkProgram now drops load errors whose file starts with another member's name. After: 0 matches.
- workspace_incremental: the runner treats names starting with `workspace_incremental` like `incremental_*` (--trace and the --full comparison). Against the committed seed's compiler (old cache, no dependents) workspace_incremental and workspace_members fail; with the branch they pass.
- item 3 (skip an unchanged-signature dependency module in a dependent program) deferred: a stale cached result for a skipped dependency module would feed the unused/effect-phase decisions of the dependent's check, so it needs more than a small change.
- mapMemberFiles maps only modules under isDependencyPath; checkWhole clears the old `<directory>\n` result keys.
