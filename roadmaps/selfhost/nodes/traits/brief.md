---
id: traits
title: Traits, constraints and equality
status: active
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
  - compiler/internal/narra/check_traits.go
  - compiler/internal/narra/emit_traits.go
  - compiler/testdata/traits*
shared:
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/types.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/emit_types.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/emit.go
reserves: []
provides: [trait-model]
depends_on:
  - on: split-compiler
    kind: hard
    why: adds trait handling to the type and expression files that split-compiler creates
---

# Traits, constraints and equality

## Goal

Add `trait`, `implements`, `implement ... for`, `derives`, constraints `<T: A & B>`, calls through the trait name, and the standard traits `Equatable`, `Hashable`, `Ordered` and `Show`. `==` then works on structs and enums. Freeze `contracts/trait-model.md`.

## Context

See `SPEC.md` section 4, "Traits", and decision D5. Implementations must be in the module that declares the type; until node modules lands, that means the same file. Generic code is compiled once per type, so constraints resolve to direct calls.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Acceptance

- [ ] Sample programs: a user trait with two implementations, a constrained generic function, derived `Show` and `Ordered`, `==` on nested structs and enums.
- [ ] Error tests: missing implementation, implementation for a type from elsewhere, `==` on a type containing a function.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: new to this repo, several syntax forms, and a contract other nodes use.

## Notes for the worker

- `Dynamic<Trait>` and `derives Shape` are not part of this node.
