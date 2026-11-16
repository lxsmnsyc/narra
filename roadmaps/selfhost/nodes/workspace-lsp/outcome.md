# Outcome: workspace-lsp

## Built

- `selfhost/src/tools/lspserver.nra`: `findPackageDirectory` skips a `narra.toml` with no `[package]` (new `isRootWithoutPackage`; a manifest that does not parse still counts as a package). `checkDocumentPackage` finds the member's workspace (`findWorkspace`, `listMembers`), builds the list of programs with `listDependentPrograms` (the member, then every member that depends on it through `workspace = true`, directly or not, in member order), and checks each through the new `checkProgram` (the old body). Each program publishes only its own modules; `Server.reported` was already keyed by package directory.
- `selfhost/src/tools/lspcache.nra`: the per-module state is keyed by package directory (`paths` and `valid` are maps by directory, `results` keys are `<directory>\n<module path>`). The parse cache stays keyed by source path, so a member shared by several programs is parsed once. Editing `app` finds `json` valid and does not check its bodies.
- `selfhost/src/tools/lspnavigation.nra`: `mapMemberFiles` maps the display file of a member's module (`json/src/parse.nra`) to its real path in the member directory. Views read those files (open documents first) and the locations use the member's URI. Before, go to definition into any dependency gave null, because `buildViews` joins the display file to the package directory.
- Cases: `testdata/lsp/workspace_incremental` (trace of alternating json and app edits), `testdata/lsp/workspace_members` (member `json`, member `app` that imports it) and `testdata/lsp/workspace_single_root` (a file under a root with no `[package]`, checked alone).

## Acceptance

- `findPackageDirectory` skips a manifest with no `[package]`; a single file under such a root is checked alone: `workspace_single_root` (publishes only for the file).
- An edit in `json` publishes diagnostics for `json` and for `app`, and `app`'s publish holds no entry for `json`'s files: `workspace_members`.
- An edit in `app` publishes `app` only and does not check `json`'s bodies again (cache keyed by directory): the last edit in `workspace_members`.
- Go to definition from `app` into `json` returns the file in the member directory: `workspace_members`, request 2.

## Deviations

- `selfhost/tests/src/lsp.nra` treats a case named `workspace_incremental*` like `incremental_*` (trace and `--full` comparison), since the runner keyed on the `incremental_` prefix.
- `lspnavigation.nra` was changed for the fourth criterion (it is in the brief's shared paths); `query.nra` stays untouched, so the member files go in through the `overrides` map that `buildViews` already takes.
- No new diagnostic codes.

## Stubs left

- Go to definition into a git dependency still gives null (the display file does not name a directory). Only members are mapped.

## Gotchas

- A body-only edit in `json` checks `json`'s bodies twice: in `json`'s program and again in `app`'s (the module's firstId changed there too). Skipping a dependency module whose signature did not change would leave a stale cached result for it in the dependent, and those results decide whether the dependent's unused and effect phases run. It needs a rule for when the stale result may stand; not built.
- A load error that lies in a member's file (a syntax error in `json`) is published by `json`'s program only; the dependent's program drops it.
- Every edit checks all dependent members; a member with many dependents costs one check each (cached per directory).
- A cycle between members is not detected here (loader gotcha).
- `closeDocument` rechecks through `checkDocumentPackage`, so it republishes the dependents too.

## Proposed learnings

- None.

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite lsp`: PASS, 21 cases.
- `sh scripts/test.sh` after the review fixes: one failure, `programs/regex_linear.nra/debug` (a wall-clock limit missed while other workers loaded the machine); it passes alone (`PASS, 2 cases`). Before the fixes: `PASS, 10290 cases in 309s`.

## Size

- Estimated M, actual M.
