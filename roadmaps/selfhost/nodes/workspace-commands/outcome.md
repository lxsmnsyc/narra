# Outcome: workspace-commands

## Built

- `selfhost/src/driver/workspaceselect.nra` (new): `selectMembers(path, dependents, dependencies)` returns every member for a workspace root, a member plus related members for `-dependents` and `-dependencies`, and None for any other path. Also `findBinaryOwners`, `checkBinaryNames` and `listMemberBinaries`. `workspace.nra` only gained `export` on `joinDirectory`, `normalizeDirectory` and `makeAbsolute`.
- `selfhost/src/driver/driver.nra`: `build`, `run`, `test` and `check` act on the selected members. `build` at the root makes every binary of every member (skips libraries; `-bin NAME` builds the owner only; `-o` needs one binary). `run -bin NAME` runs the owner (without `-bin` the workspace must have exactly one binary). `test` runs one program per member in member order and prints `== NAME (DIRECTORY) ==` before each when more than one runs; exit 1 if any fails. `check` checks each member and merges diagnostics sorted; a member's file is named from the member directory, and a diagnostic in a dependency that is a selected member is reported once, for that member. `-dependents` and `-dependencies` are flags of `build`, `test` and `check`, accepted before or after the path (after, so `test packages/json -dependents` works). `runBuild` and `runProgram` were split into `buildPath` and `runProgramCode`, which return the code, so several members run in one process.
- `selfhost/src/tools/fmtconfig.nra`: `findConfig` skips a `narra.toml` with neither `[package]` nor `[fmt]`; a package without `[fmt]` that is a workspace member takes the root's `[fmt]`.
- `selfhost/src/tools/fmt.nra`: a path that is a workspace root formats the files under every member (each file once).
- `selfhost/src/tools/query.nra`: `answerQuery` returns the answer or a `Failure`; at a workspace root every member answers and the answers are printed together. A member with no answer is skipped when another has one; a member that does not load or check fails the query. `buildViews` skips modules of dependencies (their files are not under this package; reading them failed with "No such file", also for a git dependency).
- Codes: E0488 (two members with one binary name, for `build` at the root and `run -bin NAME`) and E0489 (a flag with a path that is not a member: file, outside the workspace, or the workspace root), E0490 (a `workspace = true` cycle between members). Rows in `diagnostics.nra`.
- Docs: SPEC section 15 gets two paragraphs (workspaces; commands in a workspace) and one for `[fmt]`; D101 notes what is built; guide chapter 8 gets a Workspaces section.
- Cases: `testdata/workspaces/commands_*` (basic, check, error_binary, error_cycle, error_flags, fmt), run by the `packages` suite as `packages/workspaces/commands_*` (`runCommandsCase` in `selfhost/tests/src/packages.nra`). A case has `root/` and a `steps` file of lines `EXIT :: ARGS :: WANTS`, `exists :: PATH` and `write :: PATH :: TEXT`.

## Acceptance

- Root commands cover every member: `commands_basic` (build makes app, cli, tool; test runs four members in order; query def over members), `commands_check` (two members' errors, merged), `commands_fmt` (fmt --check at the root lists a bad file in a member and not one outside).
- `test packages/json -dependents` runs app and cli too; `check packages/cli -dependencies` covers app and json: `commands_basic`.
- `run -bin` across members: `commands_basic`. E0488: `commands_error_binary`. E0489: `commands_error_flags`, `commands_error_cycle`.
- `findConfig` skip: `commands_fmt` (`plain/deep/narra.toml`). A member without `[fmt]` takes the root's: `commands_fmt` (member `a`, tabs).
- SPEC, D101: done.

## Review fixes

- `test -filter` at the root skips a member with no matching test and fails only when no member matched.
- `check` prints a dependency's error once when several selected members report it (exact duplicates removed before sorting): `commands_check_deps`, which also checks that a selected dependency's error is named from its member folder.
- `query --json` at the root merges the members' lists into one object.
- Root `test` runs members in the order of the `members` entries (a `*` entry's matches sorted by folder): `collectMemberPaths` no longer sorts the whole list; `commands_order` has `members = ["b", "a"]`. SPEC, the guide and D101 say the same.
- `selectMembers` takes a `checksCycles` flag; `fmt` and `query` pass false, so a cycle does not stop them. A member-folder command without flags returns before listing members, and the working directory is read once.
- `check --json` steps check the codes E0489 and E0490 (E0488 comes from `build` and `run`, which have no JSON output).
- The cycle fixture uses its import. D101 lists E0480-E0490.
- After the rebase onto the landed workspace-lsp, the lsp suite passes (22 cases, including `workspace_members` and `workspace_incremental`) with the `buildViews` change.

## Deviations

- New file `selfhost/src/driver/workspaceselect.nra` and three `export` keywords in `workspace.nra`, neither in the brief's paths. Both avoid growing `driver.nra` and keep the loader's code unchanged.
- The `workspace = true` cycle between members (the loader's gotcha) is E0490, a code the coordinator added to this node. It is reported whenever a root or a flag selects members, not on a plain member build.
- `query.nra` `buildViews` skips a module of a dependency only when its file cannot be read and it has no entry in `overrides`: its file is below the dependency's name, not this package, so the read failed (a failure the root query hit). Modules present in `overrides` are used as before, so workspace-lsp can pass member files through `overrides`.
- `fmt` formats a workspace root through its members only, so files at the root outside any member are not formatted.

## Stubs left

- Workspace flags for `fmt` and `query`: not provided (the brief names only `build`, `test`, `check`).
- Language service over members: workspace-lsp. This repository as a workspace: repo-workspace.

## Gotchas

- `query` at the root names each location with the member directory (`packages/json/src/main.nra:3:1`), and `type-of` takes the same prefixed file there. A query on a member folder is unchanged.
- A root path always means every member, even when the root has a `[package]` (the root package is then one of the members).
- `fmt .` at a workspace root no longer visits directories outside the members (for this repository, `testdata/` once it is a workspace). Anything that formats those must name the paths.
- `test`, `build` and `check` options go before the path as usual; `-dependents` and `-dependencies` are also read after the path for `build` and `test`. `run` passes all later arguments to the program.
- `seed/narrac.c` is not regenerated; `selfhost/src` uses only old features.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh out`: stage 1 and stage 2 emit identical C.
- `NARRAC=out/narrac sh scripts/test.sh --suite packages --filter workspaces/commands`: PASS, 6 cases.
- Full `sh scripts/test.sh`: see the result line in the final report.

## Size

- Estimated M, actual M.

Landed on feat/batch-19 as fa449765.
