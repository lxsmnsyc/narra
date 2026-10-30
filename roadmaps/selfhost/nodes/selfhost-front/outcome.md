# Outcome: selfhost-front

## Built

- `selfhost/src/parser.nra` (module `narrac/parser`, 1501 lines). It ports `compiler/internal/narra/parser.go` plus the parse half of `check_opaque.go` (`opaqueOrDistinct`, `isWord`, `memberName`), one Narra function per Go function with the same name. Where a Go name is a Narra keyword or clashes with a variant, the Narra name differs: `is` is `at`, `next` is `advance`, `expr` is `parseExpr`, `block` is `parseBlock`, `stmt` is `parseStmt`, `pattern` is `parsePattern`, `typeExpr` is `parseType`, `decl` is `parseDecl`, `file` is `parseFile`, `arrow` is `parseArrow`. It includes PR #21's `paramAfterComma`, which also accepts `mutable` (`g: () => U uses E, mutable out: T`).
  - Public API:
    - `parse(src: string, firstId: int): Result<Parsed, Diag> uses alloc`
    - `struct Parsed { file: File, nextId: int }`
    - `struct Diag { pos: Pos, message: string }`
    - `diagText(d): string` gives `"<line>:<col>: <message>"`, the text of Go's `Diagnostic.Error()`.
  - A lex error comes back as a `Diag` too: the lexer's `"L:C: msg"` string is split, as Go's `syntaxDiagnostic` does.
- `selfhost/src/lexer.nra`:
  - The spike stubs are replaced. `unicodeEscape` builds the character with `Char.fromCode` and copies Go's `string(rune(v))`: v is cut to 32 bits, and a negative value, a surrogate or a value above 0x10FFFF becomes U+FFFD. `Char.code` replaces `charCode`. `Str.fromChar` and `Str.appendChar` replace `charString` and `pushChar`. `hex(n, width)` is now `Int.toHex` with zero padding. `KEYWORDS` (a Map) and `PUNCTS` are top-level constants.
  - `TokenKind` variants are now `TokEof TokNewline TokIdent TokInt TokFloat TokString TokChar TokTemplate TokKeyword TokPunct`, the same names as in Go.
  - `Token` gained `endLine` and `endCol`, which are Go's `Token.End`.
- `selfhost/src/main.nra`:
  - `narrac <file>` prints the token dump, as before.
  - `narrac parse <file>` prints the AST dump (`astdump.dump`) to stdout.
  - Errors (`error L:C: msg`, or a read error) go to `/dev/stderr` through `Fs.writeText`. Exit codes are 1 for an error and 2 for usage.
- `compiler/internal/narra/selfhost_test.go`:
  - `compareSelfhost` is shared by `TestSelfhostLexer` and `TestSelfhostParser`. Each test builds `selfhost/` in debug mode (ASan, UBSan, `NARRA_CHECK_LEAKS=1`). It compares stdout with the Go dump on success. On failure it expects empty stdout, the Go error line on stderr, and exit code 1.
  - Inputs:
    - every `.nra` under `compiler/testdata/` (programs, `errors/`, `panics/`, `diagnostics/`, `packages/`, `astdump/`)
    - `selfhost/src/*.nra`
    - edge cases: `lexEdgeCases` (16, including a new non-ASCII `\u{}` case) and `parseEdgeCases` (87: every parser error message, EOF inside each list loop, `describe` quoting, and syntax the samples do not use)

### Node id scheme (D63)

- One counter, `Parser.nextId`. Every `Expr` variant, every `Pattern` variant, every `Param` (function and arrow), and every `VarStmt` and `ForStmt` takes `id = nextId++` when the node is built, which is after its children. So a child usually has a smaller id than its parent.
- Ids start at the `firstId` passed to `parse`, and `Parsed.nextId` is the next free id. To number a whole package in one sequence, pass the previous file's `nextId` as the next file's `firstId`. The `narrac parse` command passes 0.
- Ids are unique within the sequence but can have gaps:
  - In `S { ... }`, the `Ident` for `S` takes an id, and the struct literal replaces it.
  - A dummy node built after an error has id -1.
  - Size side tables by `nextId`, not by the number of nodes.
- Template `${...}` expressions are parsed by a sub-parser that continues the same counter.
- Ids are not part of the dump.

### Error flow

- `Parser.failed: Option<Diag>`. `failAt(mutable p, pos, msg)` records the error only if none is set, so the first error wins, as with Go's panic. `failHere` (at the current token) and `failFound` (appends `, found <describe(tok)>`) are wrappers.
- After a failure:
  - `expect` and `ident` return the current token without advancing.
  - Parse functions return dummy nodes (`dummyExpr`, `dummyType`, `dummyBlock`, `dummyPattern`) or simply continue.
  - Every `while` loop condition starts with `ok(p)`, so all loops stop. A parse function never loops on a failed parser.
- `parse` checks `p.failed` once at the end and throws the tree away.
- A template sub-parser's error is copied to the outer parser. Its message and position are already at the template's position.
- New parse code must follow two rules: start every loop with `ok(p)`, and never index tokens past the end (`tokAt` clamps to EOF).

### Parser state

```
struct Parser {
  toks: Array<Token>   // the only array; `>>` splitting writes into it
  i: int               // current token; advance never moves past EOF
  nextId: int
  failed: Option<Diag>
}
```

Each function takes it as `mutable p: Parser`. Pure helpers take it by value: `tok`, `tokAt`, `at`, `isWord`, `ok`, `arrowAhead`, `paramAfterComma`.

## Deviations

- No file outside the owned and shared paths was edited. In the shared files, `ast.nra` has an updated comment on ids, and `astdump.nra` and `astdump.go` are unchanged: no dump mismatch came from the Go side.
- `selfhost/narra.toml` still has the comment "For now only the lexer is ported". It is not an owned path, so it was left alone.
- The harness compares stderr on failure, which the brief allows ("Update the harness if it compares stderr").
- No change requests.

## Stubs left

- `quoteGo` (Go's `strconv.Quote`, used for `found "<token>"`) and the lexer's `quoteChar` treat non-ASCII characters as printable. The exceptions are C1 controls, U+00A0, U+00AD, U+1680, U+2000-200F, U+2028-202F, U+205F-2064, U+3000 and U+FEFF, which they escape as Go does. Other characters that Go does not count as printable (unassigned code points, private use) would print differently. Only error texts are affected.

## Gotchas

- Name clashes: the parser imports `TokenKind` and the AST enums by name, so variant names must stay distinct across them. That is why the token kinds now carry the `Tok` prefix. A later module that imports both `narrac/lexer` and `narrac/ast` by name gets the same benefit.
- Mutable alias rule: `f(mutable p, g(p))` is rejected ("p is passed as mutable in this call"). Read into a local first, or use `failHere` and `failFound`. The same rule hit `Array.push(mutable lx.tokens, Token { ... lx.line ... })` in the lexer.
- Go quirks the port keeps:
  - Inner template tokens get the template's position but keep their own `End`. A string inside `${...}` dumps an `end=` relative to the inner source.
  - The template sub-parser ignores tokens after the first expression (`${a b}` parses `a`).
  - `>>` in a type is split by rewriting the current token to `>` at col+1.
  - `ifStmt` looks past newlines for `else` and rewinds if there is none.
  - `return` at EOF fails with "expected an expression, found end of file".
- `narrac parse` writes the dump with `Fs.writeText("/dev/stdout", ...)`, because `print` adds a newline and the dump already ends with one. Do not mix `print` and `/dev/stdout` writes in one command, or the order may break.
- Speed: a release build parses and dumps `parser.nra` (1500 lines, an 11k-line dump) in 0.3 s wall time.

## Proposed learnings

- Start every parser loop condition with `ok(p)` and keep the first error only in failAt | Narra has no recover; one missing check loops forever on a failed parse | selfhost/ | yes (parseEdgeCases with EOF inside each list)
- Give token kinds a `Tok` prefix | enum variants are global constructors and clash with AST variants when both are imported by name | selfhost/ | no

## Verification

- `go test ./internal/narra -run 'TestSelfhostLexer|TestSelfhostParser' -v`: 215 lexer subtests and 275 parser subtests pass. No file mismatches.
- `cd compiler && go vet ./... && go test ./...`: pass, on main 7637982 (after #21).
- `narra build -release selfhost`, then `narracr parse selfhost/src/parser.nra`: 0.3 s.

## Size

- Estimated L, actual L. The parser is 1500 Narra lines for about 1180 Go lines (1.27x), below the spike's estimate of 1800, because top-level constants and std-convert removed the workarounds.
