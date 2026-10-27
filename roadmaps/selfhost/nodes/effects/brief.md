---
id: effects
title: Effect checking
status: todo
priority: should
type: feature
outline: false
milestone: m3-core
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns:
  - compiler/internal/narra/check_effects.go
  - compiler/testdata/effects*
shared:
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/std_*.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/check_types.go
reserves: []
provides: []
depends_on:
  - on: closures
    kind: hard
    why: effect variables are written on function types, which closures adds
---

# Effect checking

## Goal

Check `uses alloc, io, unsafe` on exported functions, infer effects for private functions and arrows, and support effect variables `<effect E>` on higher-order functions.

## Context

See `SPEC.md` section 7 and decisions D8 and D14. Builtins declare their effects in the registry.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] Error test: calling an `io` function from a function that does not declare `io`.
- [ ] A generic `map` with an effect variable accepts both pure and `io` callbacks.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one analysis pass over checked code.
