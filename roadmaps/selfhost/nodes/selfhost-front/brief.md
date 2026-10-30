---
id: selfhost-front
title: Self-hosted parser
status: todo
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestSelfhostLexer|TestSelfhostParser'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - selfhost/src/parser.nra
  - selfhost/src/lexer.nra
  - selfhost/src/main.nra
  - compiler/internal/narra/selfhost_test.go
shared:
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - compiler/internal/narra/astdump.go
reserves: []
provides: []
depends_on:
  - on: selfhost-ast
    kind: hard
    why: parses into the Narra AST types and matches the dump format it defines
---

# Self-hosted parser

## Goal

Port `compiler/internal/narra/parser.go` (about 1130 lines) to `selfhost/src/parser.nra`, so that `narrac parse <file>` prints the same AST dump as `narra parse <file>` on every program in `compiler/testdata/`, including the same first error on files that fail to parse.

## Context

- Split on 2026-10-30 into selfhost-ast (dump and types, before this), this node (parser), and selfhost-modules (package loading, after this). The spike estimated the parser at about 1800 Narra lines with the error-flow cost.
- Harness: extend `TestSelfhostLexer`'s pattern in `selfhost_test.go`. Add `TestSelfhostParser`: build `selfhost/` in debug mode (ASan, UBSan, `NARRA_CHECK_LEAKS=1`), run `narrac parse` and `ASTDump` on every `.nra` under `compiler/testdata/` (programs, `errors/`, `panics/`, `diagnostics/`, packages) and on `selfhost/src/*.nra`, and compare stdout and exit code. Keep `narrac <file>` (token dump) working for `TestSelfhostLexer`; add `parse` as a subcommand.
- Node ids (D63): the parser gives every expression, pattern and local an `id: int` from a counter in the parser state. The checker keeps its results in side tables by id. Keep the scheme simple (one counter, ids unique per file or per package) and describe it in the outcome; selfhost-check-decls builds on it.
- Error flow: Go's `p.fail` panics and `file()` recovers. Keep a `failed: Option<...>` field in the parser state; once set, parse functions return dummy nodes and loops stop. Report the first error only, like Go.
- Recursion: one expression nesting level costs about 13 frames; the compiler's own source stays under 300 frames. Loop over statement and declaration lists.
- Gap workarounds are no longer needed: std-convert landed (`Char.code`, `Char.fromCode`, `Int.parse`, `Int.toHex`, `Str.endsWith`, `contains`, `replace`, `join`, NUL fix) and top-level-const landed (`const` tables, `export const`, `Alias.NAME`). Use them in the parser (`binaryLevels`, keyword sets) instead of table-building functions.
- Replace the lexer stubs from the spike in `selfhost/src/lexer.nra`:
  - `unicodeEscape` rejects escapes above 0x7F; build the character with `Char.fromCode`.
  - `charCode` (ASCII only) becomes `Char.code`.
  - `charString` and `pushChar` work around the NUL bug; use `Str.fromChar` and `Str.appendChar`.
  - `hex` and `hexDigit` become `Int.toHex` (keep the width padding).
  - `keywordTable` and `punctTable` become top-level constants.
  - `main.nra` prints errors to stdout; write them to `/dev/stderr` with `Fs.writeText` and keep the exit codes. Update the harness if it compares stderr.
- Add an edge case with a non-ASCII `\u{...}` escape to the lexer harness once `unicodeEscape` is fixed.
- selfhost-ast (#20) is merged: `ASTDump` and `narra parse` in Go, `selfhost/src/ast.nra` (module `narrac/ast`) and `selfhost/src/astdump.nra` (`dump(f: File): string uses alloc`, not yet diffed against Go). Reproduce the position rules in the "Positions" paragraph of the `astdump.go` doc comment exactly: columns count runes, a call's pos is `(`, a field-expr's pos is the member name, template parts take the template's pos.
- Go parser fix in PR #21 (open on 2026-10-30): `paramAfterComma` also accepts `mutable` after the comma, so `g: () => U uses E, mutable out: T` parses. Port the fixed version. If #21 merges while you work, rebase onto main; if it is still open when you finish, say so in the outcome.
- Read `selfhost/CLAUDE.md`, the selfhost-spike outcome (gap list, gotchas) and the selfhost-ast outcome.

## Acceptance

- [ ] `TestSelfhostParser` passes: `narrac parse` output and exit code equal `narra parse` on every `.nra` under `compiler/testdata/` and `selfhost/src/`.
- [ ] `TestSelfhostLexer` still passes, with the stubs above replaced and a non-ASCII `\u{}` edge case.
- [ ] The outcome describes the node id scheme and the error flow.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: about 1800 Narra lines of recursive descent with a hand-made error flow; one file, one harness, no new language features.

## Notes for the worker

- Keep the Go lexer's quirks (`0b102`, `3.x`, `7e`); the parser sees the same tokens.
- Newlines inside call parentheses are ignored by the lexer, so multi-statement arrow functions cannot be call arguments. Bind them to a const.
- Enum constructors through a namespace (`Ast.Expr.Call`) are not supported; import AST names directly and use `Enum.Name` on clashes.
- If a mismatch is in the Go dump (a missing field), fix `astdump.go` (shared) rather than working around it.
