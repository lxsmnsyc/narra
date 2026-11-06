---
id: mutable-bindings
title: "Mutable reference bindings and x = f(mutable x)"
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

# Mutable reference bindings and x = f(mutable x)

## Goal

Build D81 in both compilers: `mutable name = place` binds a mutable reference to a place for the rest of its block, and `x = f(mutable x)` becomes valid. The rules are in `nodes/mutable-reference/brief.md` ("Plan") and D81; follow the outcome where it changed them.

## Context

- Read `nodes/mutable-reference/outcome.md`, `nodes/closure-borrow/outcome.md` and `nodes/mutable-read-conflict/outcome.md`.
- Places and pointers: a `mutable` argument path is already lowered to a pointer after making each array on the path unique; `localRef` (`emit.go`) emits `(*name)` for `Local.ByRef` locals. A binding is a `ByRef` local initialized with that pointer.
- Conflict scans to extend: E0319 `checkExclusive`, E0929 `checkMutableConflicts` (`check_call.go`, `selfhost/src/stmts.nra`), E0915 `noMutableEscape` (`check_expr.go`, `selfhost/src/exprs.nra`), capture checks (`check_closure.go`, `selfhost/src/funcvalues.nra`).

## Acceptance

- [ ] `mutable name = place` parses as a statement in both compilers. A non-place right side (a `const`, a literal, a call) is an error with the same wording as a `mutable` argument (E0943 if a new format is needed).
- [ ] Inside the binding's block, a use of the root through an overlapping path is E0944, at the use, naming the binding and its line. Disjoint fields stay usable. Element indexes always overlap.
- [ ] Storing, returning or escaping-capturing a binding is an error (reuse E0915's rule; E0945 if the text must differ). A stack closure may write through it.
- [ ] `x = f(mutable x)` and `x.field = f(mutable x)` compile and run with the call first; other E0929 cases stay errors.
- [ ] Sample `mutable_binding.nra` and `.out`: a binding to a struct field, to an array element of a field (the array shared beforehand, so the copy is visible with `NARRA_TRACE_COPIES` or by checking the other owner), reassigning through the binding (old value released), passing it on as `mutable`, a nested block ending the borrow, a disjoint field used during the borrow, a stack closure writing through it, and `x = f(mutable x)`. ASan and leak check clean.
- [ ] Error tests `errors/mutable_binding_*.nra`, one per rule.
- [ ] Rewrite at least the five most repeated `mutable emitter.frames[...]` or `mutable checker.bodies[...]` paths in `selfhost/src` with bindings; the bootstrap fixed point holds.
- [ ] README "Differences from the spec" line from mutable-reference removed; SPEC matches. Go and `selfhost/` emit the same C; fast check passes including `TestSelfhostBootstrap`.

## Sizing

M. New statement and borrow scan in both compilers; uncertainty medium (overlap of paths), tests automatic.

## Notes for the worker

- A new statement kind must reach every walker (wave 14 lesson): exclusivity, `usesLocal`, mutable roots, narrowing ends, effects, unused-names, clone, the AST dumps, the formatter (`fmt*.nra`) and the query site index (`querysites.nra`).
- A binding is a local: unused-names and the no-shadowing rule apply to it.
