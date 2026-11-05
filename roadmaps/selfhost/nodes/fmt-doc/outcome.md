# Outcome: fmt-doc

## Built

- `selfhost/src/fmtdoc.nra`: `Doc`, `Options`, `text`, `concat`, `join`, `indent`, `group`, `print` as in `contracts/fmt-doc.md`, plus `groupWithId(contents, id)` and `runFmtDoc` (the `narrac fmt-doc <file>` command). The text form is documented at the top of the file; options are `@width`, `@tabWidth` and `@useTabs` lines.
- `selfhost/src/main.nra`: one dispatch line and import.
- `compiler/internal/narra/selfhost_fmtdoc_test.go`: `TestSelfhostFmtDoc` (goldens), `TestSelfhostFmtDocErrors`, `TestSelfhostFmtDocDeep` (50000 levels: indent, concat, group).
- `compiler/testdata/fmt/doc_*.txt` and `.out`: 14 golden cases (fit, break, nested, hardline, breakparent, ifbreak, ifbreak by id, line suffix, join, spaces, tabs, width at and over the limit, trailing whitespace).

## Deviations

- Contract interface kept. Additions: `groupWithId`. The exported `print` is a thin wrapper, because the name `print` inside the module resolves to the prelude function. The helpers `text`, `concat`, `indent`, `group` and `join` declare `uses alloc` (enum construction allocates), which the contract signatures omit.
- Output ends with a newline only when the document does.

## Stubs left

- None.

## Gotchas

- Inside `fmtdoc.nra`, `print` means the prelude function; the implementation is `printDocument`.
- fmt-cli must end its document with `Doc.HardLine` to get the final newline.
- `narra build` without `-o` only checks; use `-o` to get a binary.

## Proposed learnings

- Do the break propagation of a document printer once over an arena, not by rescanning per group | a rescan is quadratic on deep documents | selfhost/src | no

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostFmtDoc|TestSelfhostBootstrap'` and the fast check: results in the PR.

## Size

- Estimated M, actual M.
