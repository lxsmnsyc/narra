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
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: built on the self-hosted compiler (raised from later on 2026-11-01 for std-iterator)
---

# Generators with function* and yield

## Goal

`function*`, `yield`, `yield*`, `Coroutine<In, Out>`, `Co.resume`, and `for ... of` over any generator, compiled to state machines (`SPEC.md` section 8).

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
