# Outcome: selfhost-layout

## Built

- 193 modules of `selfhost/src` moved by `git mv` into `syntax/` (8), `check/` (54), `emit/` (28), `runtime/` (22), `std/` (51), `tools/` (16) and `driver/` (14). `main.nra` stays at the top. The table is in `log.md`.
- Every `narrac/<name>` path became `narrac/<group>/<name>`, in imports and in comments. `narrac fmt` re-wrapped the lines that the longer paths pushed over the width. The move script was a throwaway and is not kept.
- Test runner: `codes.nra` reads `driver/diagnostics.nra` and walks the folders (`Files.findSources`); its `localFile` rows name `syntax/parser.nra`, `syntax/lexer.nra` and `driver/manifest.nra`. `generated.nra` writes `runtime/<name>.nra` and `std/stdtext.nra`. `slotstores.nra` reads `check/slotstores.nra`.
- Paths updated in `selfhost/CLAUDE.md` (which also names the groups), `compiler/README.md`, `bench/README.md` and the roadmap contracts. `scripts/*.sh`, `.github` and `skills` name no `selfhost/src` path.
- `seed/narrac.c` regenerated in its own commit.

## Acceptance

- No `.nra` file directly in `selfhost/src/` except `main.nra`: yes.
- `sh scripts/test.sh` passes: `test.sh: PASS, 10063 cases in 210s`, on the new seed.
- No golden outside `seed/` changes: `git diff --stat feat/batch-16 -- testdata` is empty.
- `sh scripts/seed-update.sh` run twice: the second run prints `seed is current` and `seed unchanged`. `sh selfhost/bootstrap.sh`: `ok: stage 1 and stage 2 emit identical C (256018 lines)` and `seed is current`.
- `sh scripts/linux-check.sh` passes: `test.sh: PASS, 10063 cases in 201s` and `linux-check: PASS in 396s` (after the `TMPDIR` fix, see Deviations).
- `selfhost/CLAUDE.md` names the groups: yes.

## Deviations

- The first `linux-check.sh` run failed one case, `programs/windows_runtime.nra` (`createDirAll ok: false`): the program writes to `TMPDIR`, else `TEMP`, else `.`, and the container set neither, so it wrote into the read-only repository mount. This is a bug of the windows-runtime test on the base branch, not of the move. `scripts/linux-check.sh` now passes `-e TMPDIR=/tmp`, in its own commit.
- `idcheck` went to `syntax/` (it walks AST node ids) and `textops`, `hash`, `fixes` and `diagnostics` to `driver/` because every group imports them.

## Stubs left

- None.

## Gotchas

- A suite that lists `selfhost/src` by one directory level misses the folders. `codes` did; the review found that its `localFile` rows no longer matched, so the checks of the parser, lexer and manifest reporters were skipped while the suite stayed green. Fixed; a changed code in `syntax/parser.nra` makes `codes` fail again.
- A new module goes in `src/<group>/` and is imported as `narrac/<group>/<name>`.

## Contract

No contract text changed except file paths.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: `test.sh: PASS, 10063 cases in 210s`.
- `NARRAC=<built> sh scripts/test.sh --suite generated`: `PASS, 23 cases`; `--suite guide`: `PASS, 331 cases`; `--suite sources`: `PASS, 1 cases`; `--suite codes`: `PASS, 1 cases`.
- `git diff --stat feat/batch-16 -- testdata`: empty.
- Review by a fresh agent (pr-review): every one of the 193 renames equals the old file after the group prefix is removed from the import paths; one must-fix (the `localFile` rows) and one should-fix (message paths in `codes.nra`), both fixed.

## Size

- Estimated M, actual M.
