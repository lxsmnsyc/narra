# Outcome: guide-tools-gaps

## Built

- `docs/guide/10-effects-and-testing.md`: new section "Comparing text with a file" for `Test.golden`: single-file and package locations, `/` in names, the failing first run, `NARRA_UPDATE_GOLDEN=1`, the difference message, the `io` effect.
- `docs/guide/18-tools.md`: new section "Building, running and looking at the C" (`-bin`, `-o`, program arguments, `narrac emit`, `-emit-c`, `narrac targets`, a pointer to `narrac audit` in chapter 16); a note on golden files under "Running tests"; the command table now lists `build`, `run`, `emit`, `check` with `--target`, `test` with all options, `audit` and `targets`.

## Deviations

- `narrac move` is not taught. The compiler has no such command (see below), so no example can run.
- The golden example in chapter 10 is a `narra fragment` with `sh` and `text` blocks, because the guide suite cannot hold a `.golden` file. I ran it by hand with the prebuilt compiler.
- `narrac audit` is not repeated; chapter 16 already teaches it. Chapter 18 links to it.

## Spec against compiler

- SPEC.md section 3 says `narrac move old new` moves a file and rewrites imports. The compiler does not implement it: `narrac move a b` falls through to the generic usage line and exits 2 (`selfhost/src/main.nra` and `driver.nra` have no `move` case). Either implement it or remove the line from SPEC.md.
- `narrac --help` lists only build, run, test, check, targets, explain. It omits fmt, query, apply-fix, audit, emit, lsp.

## Stubs left

- None.

## Gotchas

- Node guide-targets edits chapter 18 next: it moves the build-conditions paragraph. The new "Chapter 19 builds for them" sentence in the targets section links by number only.
- Node cli-reference turns the command table into a link.

## Verification

- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide`: PASS, 525 cases.
- plain-writing checker: no warning in the new text (one older warning, line 146 of chapter 10).
- `pnpm --dir website build`: fails at `check:grammar` on chapter 2's `U64` and similar names only (known, grammar-fixes). `pnpm --dir website exec vite build`: passes.
- Examples run by hand: `test`, `NARRA_UPDATE_GOLDEN=1`, a changed golden file, a package golden under `testdata/reports/`, `build`, `build -o`, `-bin`, program arguments, `emit`, `emit --tests`, `-emit-c`, `targets`, `audit`.

## Size

- Estimated S, actual S.
