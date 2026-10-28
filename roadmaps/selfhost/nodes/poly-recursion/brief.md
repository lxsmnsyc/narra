---
id: poly-recursion
title: Reject polymorphic recursion in generic types
status: active
priority: must
type: fix
outline: false
milestone: m3-core
target:
from:
size: S
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-28
owns:
  - compiler/testdata/errors/poly_recursion*
shared:
  - compiler/internal/narra/check_types.go
reserves: []
provides: []
depends_on:
  - on: traits
    kind: hard
    why: the hang also shows in derived Show, which traits added, and the fix touches the instantiation code traits changed
---

# Reject polymorphic recursion in generic types

## Goal

The compiler must stop with an error, not loop forever, on a generic type whose instantiation never closes, such as `struct G<T> { next: Array<G<Array<T>>> }`. Each instance needs its own C type, so such a type cannot be emitted.

## Context

Found by the review of PR #3. It hangs on main before traits, and `derives Show` on the same type also hangs. Plain recursion through the same arguments (`List<T>` containing `Array<List<T>>`) must keep working.

- Source: `reviews/wave-2.md`, `SPEC.md` (generics)

## Acceptance

- [ ] `compiler/testdata/errors/poly_recursion.nra` with the type above gives a clear error at the type's declaration, and the test finishes in under a second.
- [ ] The same with `derives Show` on the type.
- [ ] Existing recursive generic programs still compile.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

S: one check in type instantiation, for example a depth limit or a cycle check over the type arguments while instantiating.
