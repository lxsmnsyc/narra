---
id: closures
title: Arrow functions and function values
status: todo
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
  - compiler/internal/langc/check_closure.go
  - compiler/internal/langc/emit_closure.go
  - compiler/testdata/closures*
shared:
  - compiler/internal/langc/ast.go
  - compiler/internal/langc/parser.go
  - compiler/internal/langc/clone.go
  - compiler/internal/langc/types.go
  - compiler/internal/langc/check_expr.go
  - compiler/internal/langc/check_call.go
  - compiler/internal/langc/emit_expr.go
  - compiler/internal/langc/emit_call.go
  - compiler/internal/langc/check_types.go
  - compiler/internal/langc/emit_types.go
  - compiler/internal/langc/check_stmt.go
reserves: []
provides: []
depends_on:
  - on: split-compiler
    kind: hard
    why: adds cases to the checker and emitter files that split-compiler creates
---

# Arrow functions and function values

## Goal

Add arrow functions, function types `(x: T) => U`, calls through function values, named functions used as values, and captures of `const` values. Function-type parameters cannot escape unless marked `escaping`, and non-escaping closures put their environment on the caller's stack.

## Context

See `SPEC.md` section 2, "Closures", and decision D46. Closures capture only constants. `let` variables and `mutable` parameters cannot be captured. Generic functions with function parameters (such as `map<T, U>`) must work.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] Sample programs cover: passing an arrow to a generic `map`, capturing strings and arrays, returning an escaping closure, storing a closure in a struct field.
- [ ] Error tests: capturing a `let`, storing a non-escaping parameter, missing `escaping`.
- [ ] Sanitizer and leak check stay clean, including for escaping closures.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: touches parser, checker and emitter, and escape checking is new.

## Notes for the worker

- `@inline` is a separate node (inline). Effect checking of function types is node effects.
