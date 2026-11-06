---
id: effects-indirect-calls
title: "Calls through struct fields and call results charge their effects"
status: todo
priority: should
type: fix
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - compiler/testdata/errors/effects_indirect_*
  - compiler/testdata/effects_indirect*
  - roadmaps/selfhost/nodes/effects-indirect-calls/**
shared:
  - compiler/internal/narra/check_effects.go
  - compiler/internal/narra/check_call.go
  - selfhost/src/effects.nra
  - selfhost/src/calls.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Calls through struct fields and call results charge their effects

## Goal

Fix an effect soundness bug on main, found in the PR #88 round-2 review (inbox, 2026-11-05). A call through a function-typed struct field or through the result of another call does not charge the effects of its function type, in both checkers:

```narra
export struct Box { run: (value: string) => int uses alloc }
export function pure(box: Box): int { return box.run("x") }   // checks clean, should be an effect error
```

A direct call to an `alloc` function in the same body is rejected correctly. Every call whose callee is a value of function type (a field, an index, a call result, a parenthesized expression) must charge the effects written on that function type, exactly as a call through a local or parameter of function type does.

## Context

- The effect pass: `check_effects.go` in Go and its selfhost counterpart (`selfhost/src/effects.nra`).
- D78 (generator effects) and the effect-variable rules (D14) say how a function-type value's effects are charged; follow the path that handles a local of function type.
- `narra query` reports `effects: []` on such calls today; after the fix it should report the charged effects. Check whether a query golden changes.

## Acceptance

- [ ] Calls through a struct field, an array index, a call result and a parenthesized function value charge the function type's effects, in Go and selfhost, with the existing effect error messages.
- [ ] Error tests `compiler/testdata/errors/effects_indirect_*.nra` cover the field, index and call-result cases; a clean sample shows the same calls in a function that declares the effects.
- [ ] Effect variables bound through such a value (a field of type `(x: int) => int uses E`) still work; add a sample if none exists.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test -timeout 45m ./...`, including `TestSelfhostBootstrap`. If `selfhost/src` itself had hidden indirect effects, add the missing `uses` there and say so in the outcome.

## Sizing

S. One rule in each checker plus tests. Risk: `selfhost/src` may contain indirect calls that now need `uses` lists.
