# Log: workspace-commands

- Selection of members in new module `driver/workspaceselect.nra` (root = every member, member + flags = related members, E0488/E0489).
- driver.nra: build, run and test dispatch over members; check merges per-member diagnostics.
- fmtconfig: findConfig skips manifests with neither table; member without [fmt] takes the root's. fmt at the root walks the members. query at the root asks every member.
- Cases: testdata/workspaces/commands_* with a `steps` file, run by the packages suite.
