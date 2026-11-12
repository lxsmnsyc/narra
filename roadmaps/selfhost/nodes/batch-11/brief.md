---
id: batch-11
title: "Batch 11: the remaining syntax features"
status: active
priority: should
type: story
story: 
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/batch-11/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-10
    kind: hard
    why: array-patterns was refined by batch-10's D103, and the D104 rule applies to the code these nodes add
---

# Batch 11: the remaining syntax features

## Goal

Eleventh one-shot batch, scope set by the maintainer on 2026-11-12: complete the language features SPEC describes that the compiler does not build, except async (batch-12), threads, async generators and C-callable functions. Children land on `feat/batch-11` as direct commits after a worker check and a fresh review. At the end: two seed updates, the full check and the Linux check, one fresh review of the whole branch, a bench row, and one pull request.

Children:

- array-patterns (L): `[]`, `[x]`, `[first, ...rest]` in `match`, `is` and `const` (D103).
- contract-logic-design (M spike, gate D102), then contract-logic (L, refined after the gate).
- generic-implement (L): constraints on struct and enum type parameters, and `implement` blocks on generic types.
- dynamic-traits (L): `Dynamic<Trait>` and `Dynamic.of`.
- distinct-traits (M): `implements` and `derives` on distinct types.
- tuple-index (S): `pair[0]` on a tuple.
- generic-field-defaults (M): defaults on struct fields whose type uses a type parameter.
- std-function-values (M): registry std functions such as `Int.toString` as values.
- inline (L): `@inline` copies the body and arrow arguments into the caller.
- lazy-values (M): `const table = lazy build()` and `Lazies.force`.
- const-eval-design (S spike, gate D105): how `@const` functions run at compile time. Its refined const-functions node joins the batch if the gate is answered in time, otherwise it waits.

const-lazy is dropped: split into lazy-values and const-eval-design.

## Order and shared files

- Machine: 8 GB, four building workers at a time. Spikes build little and run beside them.
- Order: start nodes as `plan.py ready` lists them, four building workers at a time, spikes beside them.
- Trait files (`traits.nra`, `calls.nra`, `emittraits.nra`): generic-implement first; distinct-traits and dynamic-traits start from it (hard dependencies). Its outcome records `findImplementation` and its result type.
- Emitter: std-function-values lands before inline (both change `emitIdent` and `readNodeType`). inline and lazy-values both change `emitReachableBodies`; the second to land keeps both rules and reruns the `tree-shake` emit checks.
- contract-logic waits for array-patterns (both change the `Pattern` enum and `contracts.nra`) and for the D102 gate.
- Every new syntax adds cases to the checker, emitter, fmt and astdump switches and rows to `diagnostics.nra`. Each feature's logic goes in its own module; shared files carry only dispatch. The second to land rebases and keeps both sides.
- Error codes are reserved per node: array-patterns E0722 to E0724 and E0726, tuple-index E1010 and E1011, distinct-traits E1012, std-function-values E1013, generic-implement E1020 and E1021, dynamic-traits E1022 to E1025, inline E1030 and E1031, lazy-values E1032 to E1035. Decisions: D102 contract-logic-design, D105 const-eval-design, D106 generic-implement, D108 distinct-traits.
- Each worker uses a scratch folder named after its node under /private/tmp/claude-501/. Check a rebase's exit code before testing.

## Seed

`selfhost/src` and `std/src` may use only features the committed seed compiles, so no child uses its own feature in the compiler or std. The end of the batch is two seed updates. Marking `Options.map`, `andThen` and `orElse` `@inline` comes after them, as its own step, then one more seed update.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [x] A debug-built compiler runs `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0.
- [x] `compiler/README.md` "Not supported yet" lists none of the features this batch built.
- [x] A fresh review of the whole story branch, before the pull request.
- [x] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise by more than the noise.
