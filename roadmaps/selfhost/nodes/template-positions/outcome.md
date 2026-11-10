# Outcome: template-positions

## Built

- `lexFrom` in `selfhost/src/lexer.nra` lexes a source that starts at a given line and column.
- `parseTemplate` in `selfhost/src/parser.nra` lexes each `${...}` with the position it has in the file (`advancePosition`), so every token and node inside has its real line and column. Token ends are kept in `SourceFacts.templateTokenEnds` (`selfhost/src/ast.nra`).
- The query workaround is gone: `inTemplate` on `Site` and `Use`, `TypeAt.InTemplate`, `isInterpolated`, the "not supported yet" error and the "(in a template string)" note. `querysites.nra` reads token ends from the facts and skips text-part sites inside templates.
- Samples: `testdata/errors/template_position_type.nra` and `.err`; a `type-of` query inside `${}` in `testdata/query/basic/queries.txt`.
- `compiler/README.md` drops the template limits of `type-of`, `refs`, hover, definition and references.

## Proof

- base: 6924411, `narrac check testdata/errors/template_position_type.nra` gives `3:9` for both errors (the backtick); the parts are at 3:19 and 3:38.
- branch: the same command gives `3:19`, `3:38`, `3:38`.

## Deviations

- More goldens changed than the AST dumps, because positions inside templates moved: `parse/template_nested.ast`, `parse/template_parse.ast`, `astdump/exprs.ast`, six `typesdump` goldens, four `errors/*.err`, three `lsp` goldens, two `query` goldens (the `inTemplate` field and note are gone; hover, refs and definition now answer inside `${}`). Three guide samples changed their `// error:` position (chapters 09, 10, 11).
- `facts.inTemplate` (names before a `.` in templates, used by fixes and the unused-import check) stays: it is not a position workaround.

## Stubs left

- None.

## Gotchas

- `from` is a keyword; it cannot be a parameter name.
- Text parts of a template still carry the template's position and end.

## Proposed learnings

- None.

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh`: PASS, run after the last golden update.

## Size

- Estimated M, actual M.
