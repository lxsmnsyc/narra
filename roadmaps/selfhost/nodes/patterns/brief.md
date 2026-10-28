---
id: patterns
title: Nested patterns, destructuring, spread and tuples
status: done
priority: must
type: feature
outline: false
milestone: m3-core
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns:
  - compiler/internal/narra/check_match.go
  - compiler/internal/narra/emit_match.go
  - compiler/testdata/patterns*
shared:
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/emit_types.go
reserves: []
provides: []
depends_on:
  - on: split-compiler
    kind: hard
    why: extends check_match.go and emit_match.go, which split-compiler creates
---

# Nested patterns, destructuring, spread and tuples

## Goal

Add nested patterns (`Some(Node(a, b))`), struct patterns with literal tests, `const { a, b: c } = s` destructuring, `{ ...s, field: v }` spread updates, and `Tuple<A, B>` with `[a, b]` literals and destructuring.

## Context

See `SPEC.md` section 4: Structs (Destructuring, Reading and updating), Pattern matching, and Tuples.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] Sample programs cover each form, including nested patterns with guards and exhaustiveness through nested enums.
- [ ] Error tests: field shorthand in a literal, spread of a different struct type, non-exhaustive nested match.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: several related syntax forms across parser, checker and emitter.

Pull request: #6 (merged 2026-10-28).
