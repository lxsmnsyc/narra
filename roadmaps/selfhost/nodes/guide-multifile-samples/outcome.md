# Outcome: guide-multifile-samples

## Built

- `selfhost/tests/src/guide.nra` reads a new block form, `narra file PACKAGE PATH`. Blocks of one package in a chapter add up (a later block with the same path replaces the earlier one). A file block directly followed by a `text` block builds the package so far in a temp folder, runs `narrac run` on the folder in debug (`-sanitize`) and release mode, and compares stdout. A package without a `narra.toml` block gets one with entry `main.run`. A package whose last file block has no `text` block fails the chapter (kind `dangling`); a malformed info string fails like other `narra` typos.
- Converted samples (5 run points, 10 cases): chapter 1 `shop` (narra.toml, prices, main; prints 220), chapter 4 `timeline` (span, main; default manifest), chapter 8 `words` (narra.toml, count, cli, then report and a new cli; two run points). All were `narra fragment` before.
- Docs edits: new `text` blocks after the package mains in ch1 and ch4, a `narra.toml` block in ch8, and the ch8 run instructions moved from a `sh` block into prose so the output block follows the file block directly.
- `roadmaps/selfhost/contracts/guide-samples.change-request.md` proposes contract v2.

## Proof

- base: chapter 4 fragments were never compiled.
- branch: breaking `span.start` to `span.startx` in the ch4 `src/span.nra` fragment makes `--suite guide` fail with `guide/04-structs-and-enums.md:407/debug` and `/release` ("struct Span has no field startx"). Reverted afterwards.

## Deviations

- Edited docs/guide chapters 1, 4 and 8 (shared paths), as the brief asks. Contract not edited; change request written.

## Stubs left

- None. Remaining `narra fragment` blocks are snippets in other chapters.

## Gotchas

- Plain `narra` blocks still need the `text` block right after them; so do `narra file` run points. A `sh` block between the two breaks the pairing.
- The `extractor` case does not yet cover `narra file` parsing; the real chapters do.

## Proposed learnings

- Use `narra file PACKAGE PATH` for any guide sample that spans files | untested fragments hide compile errors | docs/guide | yes (the guide suite)

## Verification

- `NARRAC=<built> sh scripts/test.sh --suite guide`: PASS, 294 cases.
- Full `sh scripts/test.sh` (built narrac): PASS, 7053 cases in 224s.

## Size

- Estimated M, actual M.
