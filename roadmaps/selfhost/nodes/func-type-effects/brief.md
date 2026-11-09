---
id: func-type-effects
title: "Function types carry their effects"
status: done
priority: must
type: fix
story: batch-1
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
spent: 252485
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

> Refined 2026-11-09 after the Go compiler was retired (PR #101): the change is in `selfhost/src` only.

## Goal

Make effect checking sound for function values. Today a function type is interned by its signature without effects, and the effect pass follows the written type of the callee (PR #98, node effects-indirect-calls). The review of #98 showed that approach cannot be made sound:

- Stores are not checked: `Plain { f: loud }` (a field typed without `uses`), `[loud]` passed as `Array<(v: int) => int>`, `function make(): (v: int) => int { return loud }` all check clean, and the later call through the slot counts as pure.
- Type parameters erase effects: `id(f)(1)`, `first(fs)(1)`, `Options.withDefault(o, f)(1)`, a generic struct field `Holder<...>.value(1)`.
- Bindings without a written type: `is Some(f)`, match arms, `for...of`, destructuring, closures returning a field, arrow parameters.

Put the effect list on the function type, as D78 did for generator types (the generator type carries its effects, binder, origin and bound, and a store into a slot with fewer effects is already rejected: `expected Generator<int>, found Generator<int> uses io`).

## Plan (D86)

- A function type's identity includes its effects. Type equality and the type key compare them; the C name of a function type ignores them, so emitted C does not change.
- Assignability: a function with fewer effects may go where more are allowed; storing one with more effects into a slot with fewer (field, element, return, argument, struct literal, `let` with a written type) is a type error naming the extra effects.
- Effect variables in function types bind as for generators (Binder and Bound).
- A call through any function value charges the effects of its type. Delete the written-type machinery from PR #98 in `selfhost/src/effects.nra`.
- `Type.TFunc` in selfhost gains the effect field (about 44 match sites).

## Acceptance

- [ ] D86 in DECISIONS.md: the rule, the subtyping direction, prior art (Koka, Effekt, D78), cost.
- [ ] Every program in `nodes/func-type-effects/review-cases/` (the #98 review's probes: stores, type parameters, bindings, closures, arrow parameters) that does an effect its function does not declare is rejected, and the clean ones stay clean, with the message text the error tests expect. Turn the distinct cases into error tests or clean samples.
- [ ] `selfhost/src` and `std/src` gain any `uses` lists the stricter rule needs; the outcome counts them.
- [ ] Emitted C unchanged for programs that check clean before and after (compare `narrac emit` output on `testdata/*.nra` from the seed compiler and the new one).
- [ ] `sh scripts/test.sh` passes, including the bootstrap suite.
- [ ] Last step: `sh scripts/seed-update.sh`, then `sh selfhost/bootstrap.sh` reports the seed is current.

## Sizing

M (roughly 300 to 500 lines in `selfhost/src` now that the Go half is gone, from the review's estimate). Risk: the stricter store rule may reject existing code in `selfhost/src` that stores effectful closures into slots typed without `uses`.

Landed on feat/batch-1 as 9e015ad.
