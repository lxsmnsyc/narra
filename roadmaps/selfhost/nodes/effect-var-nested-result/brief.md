---
id: effect-var-nested-result
title: "Effect variables leak through a generator inside another type"
status: done
priority: must
type: fix
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent: 185781
owns:
  - roadmaps/selfhost/nodes/effect-var-nested-result/**
shared:
  - selfhost/src/exprs.nra
  - selfhost/src/effects.nra
reserves: []
provides: []
depends_on: []
---

# Effect variables leak through a generator inside another type

## Goal

The checker binds a function's effect variable at a call only when the result is the generator itself (`bindCallResult` in `selfhost/src/exprs.nra`, `bindResultEffects` in `selfhost/src/effects.nra`), not a generator inside `Option`, `Array`, a tuple or a struct. So a function like `chunk<T, E>(source: Generator<T> uses E, size: int): Option<Generator<Array<T>> uses E, alloc>` lets `E` show up in the caller's types, and a private function declared `uses alloc` can run an `io` source with no error. Found by std-iterator (batch-14); the repro is `/private/tmp/claude-501/std-iterator/p4/main.nra` and the notes are in batch-14's `nodes/std-iterator/log.md` (branch feat/batch-14).

## Acceptance

- [ ] Proof: the repro on feat/batch-13 names an `E` that should not leave the callee, or misses an `io` effect.
- [ ] Fix: effect variables are bound through every generator or function type inside the result type (Option, Result, Array, tuples, structs, Map, Box, Async).
- [ ] Samples and error tests: an `io` source through an `Option`-wrapped generator is charged `io`; declaring the caller pure is an error.
