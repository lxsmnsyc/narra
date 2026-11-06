---
id: sized-f32
title: "The f32 type and float conversions"
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - selfhost/src/stdf32.nra
  - compiler/internal/narra/std_f32.go
  - compiler/testdata/f32*
  - compiler/testdata/errors/f32_*
  - roadmaps/selfhost/nodes/sized-f32/**
shared:
  - compiler/internal/narra/types.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_types.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/emit_types.go
  - compiler/internal/narra/std_float.go
  - compiler/internal/narra/std_sized.go
  - compiler/internal/narra/runtime.h
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/types.nra
  - selfhost/src/resolve.nra
  - selfhost/src/literals.nra
  - selfhost/src/exprs.nra
  - selfhost/src/traits.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/stdsized.nra
  - selfhost/src/stdfloat.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E0297]
provides: []
depends_on:
  - on: sized-conversions
    kind: hard
    why: extends its conversion table and the context typing of literals from sized-numbers in the same code
---

# The f32 type and float conversions

## Goal

Spec-gap F7, float part: `f32` in both compilers, with the rules of D82 for floats, and the conversions `F32.from`, `F32.truncate`, `F32.toString` (module `std/f32`) and `Float.from`.

## Context

- Read `nodes/sized-numbers/outcome.md` and `nodes/sized-conversions/outcome.md`. Literal typing and the no-implicit-conversion rule already exist for integers; extend them to floats.
- `f64` today: `TF64`, `Float.toString` (`lstr_from_f64`), `Float.truncate`, `Float.fromInt` (`std_float.go`).

## Acceptance

- [ ] `f32` maps to C `float`. A float literal takes `f32` when the context expects it, else `f64`; an integer literal never becomes a float. A literal outside `f32`'s finite range is E0297. `f32` and `f64` do not mix in arithmetic (E0295).
- [ ] `f32` implements `Equatable`, `Ordered` and `Show` like `f64`, and neither `Hashable` nor map keys (SPEC section 4). `Show` prints the shortest text that reads back as the same `f32`.
- [ ] `F32.from(v)` takes `f64` (rounds to nearest) or any integer type; `F32.truncate(v): int` as `Float.truncate`; `Float.from(v)` takes `f32` (exact) or any integer type. `Float.fromInt` stays.
- [ ] Sample `f32.nra` and `.out`: arithmetic, `0.1` as `f32` and as `f64` printed side by side, rounding in `F32.from`, `NaN` and infinities, a struct of `f32` fields compared with `==`. Error tests for E0297, a mixed `f32 + f64`, and an `f32` map key.
- [ ] SPEC and README updated. Go and `selfhost/` emit the same C; fast check passes.

## Sizing

M. One more primitive type in both compilers along a path sized-numbers opened; uncertainty low (shortest round-trip printing is the only tricky part).
