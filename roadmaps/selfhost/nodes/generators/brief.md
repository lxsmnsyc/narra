---
id: generators
title: Generators with function* and yield
status: todo
priority: should
type: feature
outline: true
milestone: m5-concurrency
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: split-compiler
    kind: hard
    why: adds a state-machine transform to the emitter files split-compiler creates
---

# Generators with function* and yield

## Goal

`function*`, `yield`, `yield*`, `Coroutine<In, Out>`, `Co.resume`, and `for ... of` over any generator, compiled to state machines (`SPEC.md` section 8).

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
