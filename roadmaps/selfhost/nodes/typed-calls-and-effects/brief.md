---
id: typed-calls-and-effects
title: "Effects of function values"
status: done
priority: should
type: feature
story: batch-4
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by: effect-variable-values
gate: false
added: 2026-11-04
spent: 71552
owns:
  - testdata/errors/effects_function_value*
  - testdata/effects_function_value*
  - roadmaps/selfhost/nodes/typed-calls-and-effects/**
shared:
  - selfhost/src/effects.nra
  - roadmaps/selfhost/inbox.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: checker changes go into the Narra compiler only
---

# Effects of function values

## Goal

A call through a function value that is not a plain local or parameter, such as a function-typed struct field (`box.run("x")`) or the result of a call (`make()()`), adds no effects to the caller. So `export struct Box { run: (value: string) => int uses alloc }` and `export function pure(box: Box): int { return box.run("x") }` check clean, while a direct `Str.join(...)` in the same body is rejected. Charge the declared effects of the callee's function type at every call through a value, and make `narra query` report them.

## Context

- Audit finding F20 (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md:34`), SPEC line 702, and the inbox entry from the PR #88 round-2 review (2026-11-05). Explicit type arguments (F21) already moved to node call-type-args.
- Function types already carry effects (`Type.TFunc(_, _, functionEffects, functionBinder)`, used in `addValueBoundEffects`, `selfhost/src/effects.nra:1103-1125`), and calls through locals and parameters with effect variables work (D91, node effect-variable-values). The gap is the callee forms that `scanEffects` (`effects.nra:542`) and `funcEffects` (615) do not resolve to a function type.
- Proof first: the `Box` program above checks clean on the base.
- Out of scope: a function value passed through a generic type parameter loses its effect variable binding (inbox, D91). Leave that entry.

## Acceptance

- [ ] Error tests in `testdata/errors/effects_function_value*` for a struct field call, a call result call, an array element call and a map value call, each in an exported function without the effect.
- [ ] A sample where the same calls declare the effect and run.
- [ ] Code in `selfhost/src` and `std/src` that the new check rejects is fixed (add the effect to `uses`).
- [ ] `sh scripts/test.sh` passes; the inbox entry and the README "Not supported yet" line for F20 are removed.

## Sizing

M: checker-only, but the rule may reject existing code in `selfhost/src` and `std/src` that must be fixed in the same node.
