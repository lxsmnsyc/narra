---
id: box-recursive
title: Box and recursive enums
status: active
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
  - compiler/internal/narra/emit_box.go
  - compiler/testdata/recursive*
shared:
  - compiler/internal/narra/emit_types.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/emit_match.go
  - compiler/internal/narra/runtime.h
reserves: []
provides: []
depends_on:
  - on: traits
    kind: hard
    why: moves enum payloads to the heap, which changes the equality and hash functions traits emits
  - on: patterns
    kind: hard
    why: moves enum payloads to the heap, which changes the pattern binding code patterns rewrites
  - on: split-compiler
    kind: hard
    why: changes the type layout code that split-compiler moves into emit_types.go
---

# Box and recursive enums

## Goal

Store enum variants with fields on the heap with a reference count, as `SPEC.md` section 6 says, so enums can be recursive. Add `Box<T>` with `Box.new` and `.value`.

## Context

Today enums are inline tagged unions, so `enum Expr { Add(left: Expr, right: Expr) }` is rejected as a type that contains itself. The self-hosted compiler needs recursive ASTs.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] A sample program builds and evaluates a recursive expression tree and a linked list made with `Box`.
- [ ] Leak check is clean after dropping a deep tree.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one representation change in the emitter, well covered by existing tests.

## Notes for the worker

- Enum constructors without fields can stay inline tags.
