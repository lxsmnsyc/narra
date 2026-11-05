---
id: fmt-trivia
title: "fmt: comments, blank lines and parentheses in side tables"
status: todo
priority:
type: feature
story: fmt
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostTrivia|TestSelfhostLexer|TestSelfhostParser|TestSelfhostEmit|TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - selfhost/src/trivia*.nra
  - compiler/internal/narra/selfhost_trivia_test.go
  - compiler/testdata/fmt/trivia_*
  - roadmaps/selfhost/nodes/fmt-trivia/**
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/main.nra
reserves: []
provides: [fmt-trivia]
depends_on: []
---

# fmt: comments, blank lines and parentheses in side tables

## Goal

`parseWithTrivia` and the `Trivia` side tables of `contracts/fmt-trivia.md`: the lexer keeps comments, and the parser records comment attachment, blank lines, written parentheses and node end positions, without changing `lex`, `parse`, any dump or the emitted C. The story brief (`nodes/fmt/brief.md`) has the formatter's rules.

## Context

- `selfhost/src/lexer.nra` skips comments in `lex` and emits `TokNewline` only at bracket depth 0. Tokens have `line`, `column`, `endLine`, `endColumn`.
- `selfhost/src/parser.nra`: one `Parser` struct threaded as `mutable`; ids from `Parser.nextId`; `ternaryIds` is an existing side table of the same kind. Statements, declarations and type expressions carry a `pos` but no id (`selfhost/src/ast.nra`).
- `TestSelfhostLexer`, `TestSelfhostParser`, `TestSelfhostModules` and `TestSelfhostEmit` compare the selfhost output with Go's; they must not change.
- Prettier's comment attachment (`src/main/comments/attach.js`: enclosing, preceding and following node) is the model for placement.

## Acceptance

- [ ] `narrac trivia <file>` prints the comments, their anchors and placements, `blankBefore`, `parenthesized` and `ends` in a fixed sorted text format. Golden files `compiler/testdata/fmt/trivia_<case>.nra` and `.out`, run by `TestSelfhostTrivia`, cover: leading, trailing and dangling comments on declarations, statements, struct fields, match arms, call arguments and array elements; doc comments; block comments inside expressions; blank lines between statements, declarations and fields (one and several); nested and redundant parentheses.
- [ ] For every `.nra` file under `compiler/testdata/` and `selfhost/src/`, every comment is attached exactly once, and `parseWithTrivia` returns the same AST dump and `nextId` as `parse` (checked in `TestSelfhostTrivia`).
- [ ] TestSelfhostLexer, TestSelfhostParser, TestSelfhostEmit and TestSelfhostBootstrap pass unchanged.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M. Uncertainty medium (comment attachment rules, known from Prettier), coupling low (no dependencies), blast radius medium (parser.nra is a hot file; changes are behind a flag), verifiability low (golden and equality tests), size medium (5 files, 500 to 700 lines), context medium (parser.nra is 1900 lines).

## Notes for the worker

- Collect trivia only when `parseWithTrivia` sets a flag on `Parser`; with the flag off, no side table grows, so `parse` costs the same.
- Record a node's end from the last token it consumed (`tokens[index - 1]`), at the few places nodes finish (the statement, declaration and expression parse functions), rather than adding an end position to the AST.
- Attachment can run after parsing, over the AST and the recorded starts and ends, in a separate module (`trivia.nra`), to keep parser.nra edits small.
- The Go parser is not changed.
