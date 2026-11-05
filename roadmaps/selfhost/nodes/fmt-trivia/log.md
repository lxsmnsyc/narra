# Log: fmt-trivia

- Design: the parser notes each finished node as a record (key, first token, end of last token) when `Parser.collector.enabled`. Records come in post order, so a stack pass builds a tree. Comments attach to that tree (Prettier rules) after parsing, in `trivia.nra`.
- Wrapper functions (`parseDecl`, `parseStmt`, `parseIf`, `parseExpr`, `parsePrimary`, `parseType`, `parsePattern`) keep the old bodies as `...Body`, to keep edits to parser.nra small.
- Comment types live in `trivia.nra` (not parser.nra) so lexer.nra can build them without an import cycle. The dump and `--check` live in `triviadump.nra`, which needs the lexer and parser.
- Parenthesized expressions: the record grows to include the parentheses.
- `go test` of the whole module takes minutes (ASan builds).
