# Outcome: cli-reference

## Built

- `usageText` in `selfhost/src/driver/driver.nra` lists every user-facing command with a one-line description: build, run, test, check, explain, apply-fix, fmt, query, emit, audit, lsp, targets, `--version`, `--help`. Each command line starts with `  narrac ` and carries its flags.
- Bare `narrac` prints the usage on the standard error and exits with code 2 (`selfhost/src/main.nra`). The internal dump commands (`parse`, `pkg`, `decls`, `types`, `builtins`, `trivia`, `hash`, `fmt-doc`, a bare file) still work and are not listed anywhere, not even under an "internal" heading. `commands.md` says they exist and have no stability promise.
- `docs/reference/commands.md`: a table of commands, then one section for each with its flags, what they do and its exit codes, then an environment table. Linked from `docs/reference/README.md`, replacing the "Coming" line.
- Environment variables documented: `CC`, `NARRA_HOST_TARGET`, `NARRA_CACHE`, `HOME`, `TMPDIR`/`TEMP`/`TMP`, `PWD`, `NARRA_CHECK_LEAKS` (exit 102), `NARRA_CHECK_COPIES` (exit 103), `NARRA_UPDATE_GOLDEN`, `NARRA_TEST_TIMEOUT`, and `NO_COLOR`/`FORCE_COLOR`/`TERM` for `Terminal.useColor`. Left out: `NARRA_TEST_FORCE_LEAK`, `NARRA_TEST_CHILD`, `NARRA_TEST_PIPE` (test-runner internals) and the install-script variables, which `docs/install.md` covers.
- Test `build-targets/commands-page` in the `emitchecks` suite (`checkCommandsPage` in `selfhost/tests/src/emitchecks.nra`). It runs `narrac --help`, and for each `  narrac ...` line requires `narrac <command>` and every flag (as a backtick followed by the flag) on the page. It needs only `narrac` and the repository files.
- `docs/guide/18-tools.md`: "The commands" is one paragraph naming the commands and linking the page.

## Deviations

- Bare `narrac` exits with 2 and writes to the standard error (usage error), unlike `--help`, which writes to the standard output and exits 0.
- `selfhost/tests/src/emitchecks.nra` was used for the test (not `generated.nra`), because the test runs the built compiler, as the `version` case does.

## Stubs left

- None.

## Gotchas

- The test reads a flag from the usage text and looks for a backtick followed by it. A new flag needs a row with the flag first in its table cell.
- A new command line in `usageText` must start with `  narrac ` for the test to see it.
- The seed was not regenerated. The new usage text reaches the seed with the next `sh scripts/seed-update.sh`.

## Proposed learnings

- A command or flag added to `usageText` is added to `docs/reference/commands.md` in the same change | the `build-targets/commands-page` case fails otherwise | selfhost/src/driver | yes

## Verification

- `sh selfhost/bootstrap.sh <scratch>`: fixed point holds.
- `narrac --help` and bare `narrac` print the new text.
- The `commands-page` case failed once on a usage line (`-target:` with a colon) and passed after the line was reworded.
- `pnpm --dir website build`: passes.
- `sh scripts/test.sh`: PASS, 10736 cases in 323s.

## Size

- Estimated M, actual M.
