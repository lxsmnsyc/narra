---
id: selfhost-spike
title: Self-hosting gap check
status: todo
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: true
verify: cd compiler && go test ./internal/narra -run TestSelfhostLexer
exclusive: []
fulfilled_by:
gate: true
added:
owns:
  - selfhost/**
  - compiler/internal/narra/selfhost_test.go
  - compiler/internal/narra/lexdump.go
shared:
  - compiler/cmd/narra/main.go
reserves: []
provides: []
depends_on:
  - on: modules
    kind: hard
    why: the compiler port is a multi-module package
  - on: maps
    kind: hard
    why: the port needs maps for symbol tables
  - on: strings-std
    kind: hard
    why: the lexer needs string positions
  - on: io-std
    kind: hard
    why: the driver reads files and runs the C compiler
  - on: box-recursive
    kind: hard
    why: the AST is recursive
  - on: patterns
    kind: hard
    why: the checker relies on nested patterns
  - on: closures
    kind: hard
    why: the port uses higher-order helpers
  - on: traits
    kind: hard
    why: symbol tables need Hashable keys
---

# Self-hosting gap check

## Goal

Port the bootstrap lexer (`compiler/internal/narra/lexer.go`, about 370 lines) to Narra as the first module of the self-hosted compiler, and prove it produces the same tokens as the Go lexer on every program in `compiler/testdata/`. While porting, list every language or std feature the rest of the compiler (parser, checker, emitter, driver) will need that the bootstrap compiler lacks, and estimate the port. The user reviews that list before selfhost-front starts.

## Context

- The self-hosted compiler lives in `selfhost/` at the repo root, a package with `narra.toml` (see the modules outcome) and modules under `selfhost/src/`. This node creates the package and `src/lexer.nra` plus a small `src/main.nra` that reads a file and prints its tokens.
- Token dump format: one token per line, `<line>:<col> <kind> <text>`, with text escaped so newlines and tabs stay on one line. Add the same dump to the Go side as a hidden command `narra lex <file>` (code in `lexdump.go`, one `case` in `cmd/narra/main.go`), so both sides print the same thing.
- `selfhost_test.go` builds `selfhost/` with the Go compiler, runs it and `narra lex` on every `.nra` file under `compiler/testdata/` (including `errors/`, where lexing should still succeed or fail the same way), and diffs the output. The built program runs under AddressSanitizer and the leak check like the other samples.
- Available now: maps (`Map<K, V>`, if #13 has merged; otherwise use arrays and note it), strings-std (`Str.slice`, `charAt`, `nextChar`, `StringBuilder`; positions are plain `int` byte offsets, slices are copies), io-std (`std/fs`, `std/process`, `std/env`, `Result` errors), `Box<T>` and recursive enums, closures, traits, nested patterns, modules, effects (exported functions declare `uses`). Read `compiler/README.md` for what is supported and the differences from the spec.
- User-code recursion overflows the C stack at about 20000 frames. The lexer must loop, not recurse per character or token. Note in the gap list whether the parser's recursive descent stays within that for the largest input it will see (the compiler's own source).
- Generated C names are hashed (xxHash32), so the emitter port needs wrapping `*` and bit rotates on 32-bit integers. Check whether `int` arithmetic gives that and list it if not.

## Acceptance

- [ ] `selfhost/` builds with `narra build`, and its lexer output matches `narra lex` on every `.nra` file under `compiler/testdata/` (`TestSelfhostLexer`).
- [ ] `outcome.md` holds the gap list: each missing feature with the compiler part that needs it, a workaround if one exists, and a size guess. Group by blocker (port cannot proceed), annoyance (workaround costs code or speed) and nice-to-have.
- [ ] `outcome.md` holds a port estimate per remaining node (selfhost-front, selfhost-check, selfhost-emit, selfhost-bootstrap) in source lines and size, and proposes new nodes for the blockers.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one small module ported line by line, a test harness and a written report. The gap list is the deliverable that matters.
