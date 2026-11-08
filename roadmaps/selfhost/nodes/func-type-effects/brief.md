---
id: func-type-effects
title: "Function types carry their effects"
status: todo
priority: must
type: fix
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - testdata/errors/func_effects_*
  - testdata/func_effects*
  - roadmaps/selfhost/nodes/func-type-effects/**
shared:
  - selfhost/src/**
  - DECISIONS.md
  - SPEC.md
  - compiler/README.md
  - testdata/**
reserves: [decision D86]
provides: []
depends_on:
  - on: effects-indirect-calls
    kind: hard
    why: replaces its written-type machinery in check_effects.go and effects.nra
  - on: c-seed
    kind: hard
    why: user decision 2026-11-08, wait for the Go compiler to be retired so the fix is built once, in Narra only
---

# Function types carry their effects

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

Make effect checking sound for function values. Today a function type is interned by its signature without effects, and the effect pass follows the written type of the callee (PR #98, node effects-indirect-calls). The review of #98 showed that approach cannot be made sound:

- Stores are not checked: `Plain { f: loud }` (a field typed without `uses`), `[loud]` passed as `Array<(v: int) => int>`, `function make(): (v: int) => int { return loud }` all check clean, and the later call through the slot counts as pure.
- Type parameters erase effects: `id(f)(1)`, `first(fs)(1)`, `Options.withDefault(o, f)(1)`, a generic struct field `Holder<...>.value(1)`.
- Bindings without a written type: `is Some(f)`, match arms, `for...of`, destructuring, closures returning a field, arrow parameters.

Put the effect list on the function type, as D78 did for generator types (`GeneratorType` has `Effects`, `Binder`, `Origin`, `Bound`, and a store into a slot with fewer effects is already rejected: `expected Generator<int>, found Generator<int> uses io`).

## Plan (D86)

- A function type's identity includes its effects. `sameType` and the type key compare them; the C name (`funcTypeCName`) ignores them, so emitted C does not change.
- Assignability: a function with fewer effects may go where more are allowed; storing one with more effects into a slot with fewer (field, element, return, argument, struct literal, `let` with a written type) is a type error naming the extra effects.
- Effect variables in function types bind as for generators (Binder and Bound).
- A call through any function value charges the effects of its type. Delete the written-type machinery from #98 (`typeExprOf`, `writtenEffects`, `resultEffects` and the selfhost helpers).
- `Type.TFunc` in selfhost gains the effect field (about 44 match sites).

## Acceptance

- [ ] D86 in DECISIONS.md: the rule, the subtyping direction, prior art (Koka, Effekt, D78), cost.
- [ ] Every program in `nodes/func-type-effects/review-cases/` (the #98 review's probes: stores, type parameters, bindings, closures, arrow parameters) that does an effect its function does not declare is rejected, and the clean ones stay clean, in both compilers with the same text. Turn the distinct cases into error tests or clean samples.
- [ ] `selfhost/src` and `std/src` gain any `uses` lists the stricter rule needs; the outcome counts them.
- [ ] Emitted C unchanged for programs that check clean before and after (compare `narrac emit` on the test programs).
- [ ] Fast check passes, including `TestSelfhostBootstrap`.

## Sizing

M (roughly 600 to 1,000 lines across both compilers, from the review's estimate). Risk: the stricter store rule may reject existing code in `selfhost/src` that stores effectful closures into slots typed without `uses`.
