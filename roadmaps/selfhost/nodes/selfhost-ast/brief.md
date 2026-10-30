---
id: selfhost-ast
title: AST dump and Narra AST types
status: todo
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestASTDump|TestSelfhostLexer'
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns:
  - compiler/internal/narra/astdump.go
  - compiler/internal/narra/astdump_test.go
  - compiler/testdata/astdump/**
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
shared:
  - compiler/cmd/narra/main.go
reserves: []
provides: []
depends_on:
  - on: selfhost-spike
    kind: hard
    why: extends the selfhost package and its dump harness pattern
---

# AST dump and Narra AST types

## Goal

Define a fixed text dump of the parsed AST, print it from the Go parser as a hidden command `narra parse <file>`, and write the Narra AST types and the same printer in `selfhost/`. selfhost-front then ports the parser and diffs its dump against the Go one.

## Context

- Split from selfhost-front on 2026-10-30: the spike estimated the front end at 3500-4000 Narra lines, more than one L node.
- Follow the token dump from selfhost-spike: `compiler/internal/narra/lexdump.go` (`LexDump`) and the `lex` case in `compiler/cmd/narra/main.go`. Add `ASTDump(src) (string, bool)` in `astdump.go` and a hidden `parse` case.
- Format: one node per line, indented by depth, `<kind> <fields>` with positions `line:col`. Every field the checker reads goes in (names, operators, literal text, `export`, `opaque`, `mutable`, type expressions, `uses`, effect parameters). Escape text like `escapeText` in `selfhost/src/lexer.nra`. A parse error prints `error <line>:<col>: <message>` and exits 1, as `narra lex` does. Do not print node ids; they are internal to the Narra parser.
- Narra types in `selfhost/src/ast.nra`: one enum variant or struct per Go type in `compiler/internal/narra/ast.go` (48 structs today, including `ConstDecl`, `DistinctDecl`, `MapLit`). Every expression, pattern and local carries `id: int` (D63). Children are values inside `Box` or arrays; nothing points back.
- `selfhost/src/astdump.nra` prints the Narra AST in the same format.
- Read `selfhost/CLAUDE.md` first.

## Acceptance

- [ ] `narra parse <file>` prints the dump; `TestASTDump` runs it on every `.nra` under `compiler/testdata/` (no panic, same output twice) and compares a few files with goldens in `compiler/testdata/astdump/` that together cover every AST node type.
- [ ] `selfhost/src/ast.nra` has a type for every node in `ast.go`, and `selfhost/` still builds (`TestSelfhostLexer` passes).
- [ ] A doc comment at the top of `astdump.go` defines the format, so selfhost-front can match it.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: a Go printer over 48 node types, the same types in Narra, and a Narra printer. No parsing logic.

## Notes for the worker

- The Narra printer cannot be compared with the Go one until the parser exists. Write it carefully against the format comment; selfhost-front fixes any mismatch.
- Keep the dump stable: map fields and anything sorted must print in a fixed order.
