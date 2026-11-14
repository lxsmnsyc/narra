# Outcome: field-shorthand-sweep

## Built

- Every `name: name` struct field in `selfhost/src/*.nra` and `std/src/*.nra` is now the shorthand `name`: 563 sites in 73 files (the brief guessed 512). Done by one regex pass over the sources, then `narrac fmt`; nothing by hand. `stdtext.nra` regenerated with `sh scripts/test.sh --suite generated --update`; `seed/narrac.c` regenerated twice (the second run printed `seed unchanged`).
- `fmt` joined one call in `lexer.nra` onto one line, so four lines moved there.

## Proof

- base: 034ef48f, a scan for struct fields `name: name` (comments skipped) over `selfhost/src` and `std/src` finds 563 sites.
- branch: the same scan finds 0.

## Deviations

- None. `stdtext.nra` and `seed/narrac.c` changed only through their commands.

## Stubs left

- None.

## Gotchas

- `stdtext.nra` embeds the std sources, so the C of the whole tree differs before and after by those string literals. Compare with the swept `stdtext.nra` placed in both trees.
- Emitted bounds-check strings carry source positions (`src/lexer.nra:LINE:COL`); any fmt line join shifts them.
- A field written `name: name` in new code is not an error; fmt keeps the form as written.

## Proposed learnings

- Use the shorthand when a field's value is a local of the same name | the sweep made it the style of the compiler | selfhost/src, std/src | no

## Verification

- Same compiler (built from feat/batch-15), `narrac emit selfhost` of the unswept tree (with the swept `stdtext.nra`) against the swept tree: 246442 lines each, identical except 26 `src/lexer.nra:LINE:COL` strings shifted by the fmt join; identical when those are masked.
- Compiler built from the branch seed against the base compiler, same source: byte-identical C, for both the swept and the unswept tree. So code that does not use the feature emits the same C.
- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 C identical; `sh scripts/seed-update.sh` twice: `seed updated`, then `seed current`, `seed unchanged`.
- `narrac fmt --check selfhost/src std/src` clean. Full `sh scripts/test.sh`: see the report.

## Size

- Estimated M, actual S.
