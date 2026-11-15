# Outcome: guide-tools

## Built

- `docs/guide/12-tools.md`, chapter 12 "Tools": `narrac check` (exit codes, `--all-targets`, `--json` and its fields), `narrac explain`, `narrac apply-fix` (`--diagnostic`, `--fix`, saving the result), `narrac fmt` (`--check`, `--stdin`, the `[fmt]` table with its four keys and defaults), `narrac query` (all six questions, `--package`, `--json`, exit codes), `narrac test` on a package, and `narrac lsp` (features, Neovim and Helix setup, `--full`, `--trace`). It builds one package, `weather`, as `narra file` blocks that the guide suite runs. Every shell session output was taken from a narrac built from this branch.
- The TypeScript comparison moved to `docs/guide/13-narra-for-typescript-programmers.md` (heading renumbered). It gained a short "Tools" section (tsc, Prettier, Jest, the TypeScript language server), so the tools chapter has no comparisons.
- `docs/guide/README.md`: chapter list has 12 and 13. "Planned" now lists only what is not built: package dependencies and workspaces, C-callable functions and `@layout`, Windows and WebAssembly targets. Async, threads, SIMD and `@const` moved to a new list of built features the guide does not teach yet, with fixed arrays, `ensures` and refinement types, `typeof`, `Dynamic`, unsafe code, C interop and build conditions.
- `docs/guide/11-standard-library.md` links to chapter 12; `docs/guide/01-getting-started.md` points to chapter 12 after its command table.

## Deviations

- Edited `docs/guide/11-standard-library.md` (its "Next" link pointed at the renamed file). Not in owns or shared; one line.
- `docs/guide/13-narra-for-typescript-programmers.md` is a rename of the shared `12-...` file, plus the new Tools section.

## Stubs left

- None.

## Gotchas

- The guide suite does not check links. I checked them with a script (relative links and `#anchors` in docs/guide: 0 bad). A link check in `selfhost/tests/src/guide.nra` would keep it so.
- `narrac check` text output has no diagnostic code; only `--json` shows it, so a user who wants `narrac explain` needs the JSON form. The chapter says so.
- E0800 is shared by several messages (missing `uses` on an exported function and the trait implementation rule). `narrac explain E0800` prints the last format, the trait one, so it does not match the common missing-`uses` error. The chapter uses E0003 instead.

## Proposed learnings

- Guide shell sessions should be captured from a built narrac, not written from the source comments | outputs such as `fmt --check` paths and query text are easy to guess wrong | docs/guide | no

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh --suite guide`: PASS, 328 cases.
- Link check script over docs/guide: 0 bad links.
- plain-writing `check_prose.py docs/guide/12-tools.md`: ok.
- Fast check `sh scripts/test.sh`: PASS, 9781 cases (a first run at load 33 failed only the timing case programs/regex_linear.nra/debug, "over 1 s"; it passes alone and in the rerun).

## Size

- Estimated M, actual M.
