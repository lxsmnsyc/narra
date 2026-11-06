---
id: mutable-bindings
title: "Build D81: mutable changes contents only"
status: todo
priority: should
type: feature
story:
early: false
outline: true
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
  - compiler/testdata/mutable_binding*
  - compiler/testdata/errors/mutable_binding_*
  - roadmaps/selfhost/nodes/mutable-bindings/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_closure.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/emit.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/stmts.nra
  - selfhost/src/exprs.nra
  - selfhost/src/funcvalues.nra
  - selfhost/src/scopes.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/fmt.nra
  - selfhost/src/querysites.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E0943, diag E0944, diag E0945]
provides: []
depends_on:
  - on: mutable-reference
    kind: hard
    why: the gate fixes the rules (D81) this node builds
  - on: closure-borrow
    kind: hard
    why: stack closures capture a reference binding by reference, through the capture mode and conflict scan closure-borrow adds
---

# Build D81: mutable changes contents only

## Goal

Outline, refined after mutable-reference passes its gate. Build D81 in both compilers: `mutable` parameters only of struct, `Array` or `Map` type, no assignment to the whole parameter, `const` bindings passable as `mutable` and with assignable fields and elements. Rewrite every place the mutable-reference outcome counts, in `selfhost/src`, `std/src`, `compiler/testdata` and `docs/guide`. Bootstrap fixed point holds.

The first draft of this node (local reference bindings `mutable name = place` and `x = f(mutable x)`) is parked in the inbox; see mutable-reference's brief.

Size guess: M to L, mostly the rewrite of existing code.
