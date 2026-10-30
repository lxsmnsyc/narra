# Log: selfhost-front

- 2026-10-30: branch feat/selfhost-front from origin/main c5c7dbd. PR #21 merged during the work; fast-forwarded to 7637982 (before the first commit), so `param_after_uses.nra` is covered.
- Lexer: TokenKind variants renamed to `Tok*` (as in Go) so they do not clash with the `Expr` variants `IntLit`, `Ident`, ... when both are imported by name. Token gained `endLine`/`endCol` (Go `Token.End`), needed for `StringLit.end`. Stubs replaced: `Char.fromCode` (with Go's `string(rune(v))` rules: 32-bit cut, invalid gives U+FFFD), `Char.code`, `Str.fromChar`/`Str.appendChar`, `Int.toHex` in `hex` (padding kept), `KEYWORDS`/`PUNCTS` constants.
- Parser: one function per Go function. `failed: Option<Diag>`; every loop condition starts with `ok(p)`. Template `${}` uses a sub-Parser sharing the id counter; its error is copied to the outer parser.
- Mutable-alias rule: `failAt(mutable p, posOf(tok(p)), ...)` is rejected (p used twice in one call). Added `failHere`/`failFound` that read the token first.
- `quoteGo` copies strconv.Quote; non-ASCII "not printable" handled only for C1 controls, NBSP, soft hyphen and the U+2000 block spaces/format chars, U+3000, U+FEFF.
- main.nra: errors go to /dev/stderr; the dump goes to /dev/stdout through `Fs.writeText` (print would add a newline). Harness compares stdout on success and stderr on failure.
- Harness: shared `compareSelfhost`; 87 parse edge cases (every parser error, EOF inside each list loop, describe quoting) and a non-ASCII `\u{}` lex case.
- Result: all 275 parser subtests and 215 lexer subtests pass first try after compile fixes. No mismatching files.
