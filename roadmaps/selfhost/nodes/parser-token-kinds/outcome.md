# Outcome: parser-token-kinds

## Built

- `selfhost/src/lexer.nra`: `Token` has a new field `symbol: int`. `emit` sets it from the const table `SYMBOLS_BY_TEXT` for keyword and punctuation tokens (0 for every other token). The 66 keywords and punctuation strings the parser tests are exported as `SYMBOL_*` constants (`SYMBOL_LEFT_PAREN`, `SYMBOL_KEYWORD_IF`, ...). Token text and kinds are unchanged, so `narrac lex` dumps are unchanged.
- `selfhost/src/parser.nra`: `isAt`, `accept`, `expect` and `parsePatternList` take a symbol instead of text. `isAt` compares `parser.tokens[parser.index].symbol`. `BINARY_LEVELS` holds symbols and `parseBinary` compares `token.symbol`, including the `===` and `!==` check. `getSymbolText` finds the text for the `expected "..."` message in `expect`. The `>>` split keeps `symbol: SYMBOL_GREATER`.
- Other `token.text ==` tests in the parser (identifiers like `this`, `new`, `class`, lookahead helpers) are untouched.

## Proof

- base: cc7baebe, `/usr/bin/time -l narrac emit selfhost`, instructions retired, 8 alternating pairs, median 1809.5M (range 1793.6M to 1828.2M, 1.9%). Load average 22 to 33.
- branch: same command, median 1754.5M (range 1748.5M to 1782.6M, 2.0%). Gain 3.0%; the branch won 8 of 8 pairs, the smallest pair gain is 2.2%. The gain is larger than the 1.5% the profile predicted and above the run-to-run spread.
- Both compilers were built with `sh selfhost/bootstrap.sh` and emit identical C for `selfhost` (`cmp`).

## Result

Not an experiment node.

## Deviations

- Symbols are plain `int` constants in a table, not a new enum, so call sites compare ints without enum equality. The table holds only texts the parser tests, so the constants and the table are one list in `lexer.nra`.

## Stubs left

- None.

## Gotchas

- A new keyword or punctuation the parser tests needs a `SYMBOL_` constant and a `SYMBOLS_BY_TEXT` row; a token with no row has symbol 0 and never matches.
- The seed was not regenerated. Measurements use compilers built by `sh selfhost/bootstrap.sh`.
- body-id-ranges edits parser.nra after this node; the call sites now read `isAt(parser, SYMBOL_...)`.

## Proposed learnings

- Compare tokens by int symbol, not by text | text compares cost memcmp on every parser test | selfhost/src | no

## Verification

- `sh selfhost/bootstrap.sh`: fixed point holds (stage 3 path, seed older than source, as on the base).
- `sh scripts/test.sh`: result in the worker report.

## Size

- Estimated M, actual M.
