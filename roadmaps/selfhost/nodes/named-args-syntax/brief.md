---
id: named-args-syntax
title: Parse named arguments
status: active
priority:
type: feature
story: batch-1
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - testdata/parse/named_args*
  - testdata/fmt/named_args*
  - roadmaps/selfhost/nodes/named-args-syntax/**
shared:
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmtexprs.nra
reserves: []
provides: []
depends_on: []
---

# Parse named arguments

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

`Arg` gets an optional name, in both compilers. An argument that starts with a name followed by `:` is named: `move(x: 1, y: 2)`, `fill(buffer: mutable xs, value: 0)`. The AST keeps the written order and the name's position. astdump prints the name, `clone.go` copies it, and `narrac fmt` prints `name: value` with the existing argument layout. Until named-args-check lands on the story branch, the checker rejects a named argument with a plain error at its position (E0000 is fine on the story branch; it never reaches main).

## Context

- `callArgs` in `parser.go` (about line 1138) and its selfhost twin. No expression starts with `name :`: an arrow starts with `(`, and a conditional's `:` follows a `?`. A block argument `{ name: for ...}` is inside braces and is not affected.
- D87 and `roadmaps/selfhost/debate/named-arguments/debate.md` for the rules; this node only parses.

## Acceptance

- [ ] Parse and astdump goldens for named, mixed and `mutable` named arguments, equal in Go and selfhost.
- [ ] A fmt golden: named arguments keep their order and spacing, and a long call breaks one argument per line as positional calls do.
- [ ] Fast check passes.
