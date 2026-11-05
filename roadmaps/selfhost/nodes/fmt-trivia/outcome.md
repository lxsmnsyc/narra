# Outcome: fmt-trivia

## Built

- `selfhost/src/trivia.nra`: `CommentKind`, `Placement`, `Comment`, `Attached`, `Trivia` (as in the contract), the parser's `Collector`, anchor key helpers (`getPosKey`, `getIdKey`), and `buildTrivia` (tree of node records, Prettier-style comment attachment, blank lines, ends), `checkAttached`.
- `selfhost/src/triviadump.nra`: text format of `narrac trivia <file>` and `narrac trivia --check <file>` (compares `parseWithTrivia` with `parse`: tokens, tree dump, nextId or error, every comment attached once).
- `selfhost/src/lexer.nra`: `lexWithComments` (same tokens as `lex`, plus comments). `lex` unchanged in output.
- `selfhost/src/parser.nra`: `parseWithTrivia(source, firstId): Result<#[Parsed, Trivia], Diagnostic>`; node notes behind `Parser.collector.enabled`.
- `selfhost/src/main.nra`: the `trivia` command.
- `compiler/internal/narra/selfhost_trivia_test.go` (`TestSelfhostTrivia`), goldens `compiler/testdata/fmt/trivia_{comments,blank,parens}.nra/.out`.

Dump format, one line each: `comment <i> <Line|Block|Doc> l:c-l:c <own|inline> "<text>"`, `attached <key> <leading|trailing|dangling> <i>`, `blank <key>`, `paren e<id>`, `end <key> l:c`; sorted by key.

## Deviations

- Types `CommentKind`, `Placement`, `Comment`, `Attached`, `Trivia` are exported from `narrac/trivia`, not `narrac/parser` (import cycle with the lexer). `parseWithTrivia` is in parser.nra as in the contract. Consumers import the types from `narrac/trivia`.
- Extra anchor kind `b<line>:<col>` for blocks (the `{` position), so comments in an empty block are dangling on the block. Statements nest under it.
- Extra blankBefore keys `c<index>` for comments with a blank line before them. Interpretation of blankBefore: a node has it when a blank line lies between it and what comes before it, which is its last leading comment, else the previous sibling (including that sibling's trailing comments). So a first item has it only when a leading comment precedes it with a blank line between. blankBefore is set only for statements, declarations, fields, variants, arms and imports (not for expressions).
- The `ends` and tree ranges of a parenthesized expression include its parentheses. `parenthesized` has expression ids only; patterns are never parenthesized in Narra.
- Anchor keys use the node's `pos` field (declaration pos is the keyword after `export`; a labeled loop uses the label's pos; `return c ? a : b` uses the pos of the converted `if`). Nodes the parser makes up (ternary and for-pattern desugaring) have no record.
- Not recorded: type-parameter bounds, import names, map entries, call arguments (the argument expressions are recorded). Comments inside `${...}` of templates are not collected (they are lexed on their own). A comment in an empty call `f(/* c */)` is trailing on the callee, not dangling.
- No change request file: the contract text still holds for what it states.
- Files touched outside owned paths: none besides the shared lexer.nra, parser.nra, main.nra.

## Stubs left

- None.

## Gotchas

- Edits in parser.nra: new wrappers around `parseDecl`, `parseStmt`, `parseIf`, `parseExpr`, `parsePrimary`, `parseType`, `parsePattern` (bodies renamed `...Body`); `parsePostfix` gained a `startIndex` parameter; the `Parser` struct gained `collector` (both construction sites). Merge conflicts with the `function*`/`yield` change would be in those spots.
- `parseLabeledStmt` calls `parseStmtBody` so only the labeled statement is recorded.
- A new AST node that should take comments needs a `noteKey`/`noteExpr` call where it is built.
- The inner parser of template strings has the collector disabled.

## Proposed learnings

- Keep optional collection in the parser behind one `enabled` flag on a side struct | `parse` costs and output stay the same | selfhost/ | yes (TestSelfhostTrivia --check)

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostTrivia` passes (3 goldens, `--check` on every .nra under testdata/ and selfhost/src plus three edge inputs).
- Fast check `cd compiler && go vet ./... && go test ./...`: see the PR.

## Size

- Estimated M, actual M.
