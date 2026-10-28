---
id: maps
title: Map with insertion order
status: todo
priority: must
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
  - compiler/internal/narra/std_map.go
  - compiler/internal/narra/runtime_map.h
  - compiler/testdata/maps*
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/emit_types.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
reserves: []
provides: []
depends_on:
  - on: traits
    kind: behavior
    via: trait-model
  - on: split-compiler
    kind: contract
    via: builtins
---

# Map with insertion order

## Goal

Add `Map<K, V>` as `SPEC.md` describes: `Map { "a": 1 }` literals, `m[k]` returning `Option<V>`, assignment, `Map.keys`, `Map.entries`, `Map.size`, `Map.remove`, insertion order, and copy on write.

## Context

Keys need `Hashable` and `Equatable` from `contracts/trait-model.md`. Until traits lands, build against string and integer keys and the contract's function names. Iteration with `for ... of` can be special-cased in the emitter, as `Array.values` is, until generators lands.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

- From split-compiler: put new runtime C in the module's `Module.Runtime` (in its `std_*.go` file), not in `runtime.h`. The emitter writes it once per imported module.

## Acceptance

- [ ] Sample program counts words with a `Map<string, int>` and prints them in insertion order.
- [ ] Error test: a map with a float key.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one new container following the array's pattern.
