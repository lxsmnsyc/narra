# Outcome: lexer-slices

## Built

- `selfhost/src/lexer.nra` reads the source string by byte (`Str.byteAt`) and cuts token text with `Str.slice`. `Lexer.source` is a `string`, `Lexer.index` a byte offset, and the `punctuationChars` field is gone.
- Columns still count characters (as before): `advance` skips the column step for UTF-8 continuation bytes (`0x80` to `0xBF`).
- A quoted literal without escapes is one slice. With escapes, runs between them are sliced and appended to a builder. Numbers are one slice, with `_` removed only when present.
- Punctuation is matched against the `PUNCTS` strings by byte, and advances by the punctuation's length.
- Public functions are unchanged (`lex`, `lexFrom`, `lexWithComments`, `decodeEscape(Array<char>, ...)`, `Escape`, `Token`, `LexError`, `hex`, `escapeText`, `getKindName`, `formatLexError`). `decodeEscape` no longer builds a `Lexer`; it decodes on the `Array<char>` directly and reports the same codes and positions. `parser.nra` is not touched.
- New golden `testdata/lex/bad_after_multibyte` (error column after a `é` and a `✓`).

## Proof

- base: feat/batch-4 at ffdea9c, `/usr/bin/time -l narrac emit selfhost`, 3 runs: instructions retired 9.399e9, 9.355e9, 9.319e9 (spread 0.9%); peak 130.5, 135.9, 135.4 MB.
- branch: 8.403e9, 8.400e9, 8.370e9 (spread 0.4%); peak 134.9, 137.3, 135.2 MB. About 10.5% fewer instructions, well outside the spread. Peak memory is within the noise.
- Same token output: `narrac <file>` (lex dump) of the base and branch compilers on all 1270 `.nra` files under `testdata`, `selfhost` and `std`, plus 7 extra files with multi-byte characters before errors, escapes, numbers and comments, differ in none.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- `decodeEscape` still takes `Array<char>` because `parser.nra` (parseTemplate) passes the characters of the template. A later change that makes the parser slice template text can drop it and use the byte version.
- Index inside the lexer can sit in the middle of a multi-byte character after an escape or `\` skip; only slice ends at ASCII delimiters are used, so slices stay on character boundaries.

## Proposed learnings

- Lexer columns count characters, so a byte-based `advance` must not step the column on continuation bytes | the LSP converts these columns to UTF-16 | selfhost/src | yes (testdata/lex/bad_after_multibyte)

## Verification

- `sh selfhost/bootstrap.sh <dir>` on the branch: stage 1 and stage 2 emit identical C (184138 lines).
- `sh scripts/test.sh`: see the orchestrator report.

## Size

- Estimated M, actual M.
