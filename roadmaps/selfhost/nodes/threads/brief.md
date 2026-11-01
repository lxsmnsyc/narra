---
id: threads
title: Scoped threads, channels and shared state
status: todo
priority: later
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
  - on: closures
    kind: hard
    why: thread functions are escaping closures
  - on: identity-cycles
    kind: hard
    why: Shared, channels and tasks follow the cycle decision
---

# Scoped threads, channels and shared state

## Goal

`Thread.scope`, `Thread.start`, channels, `Shared<T>`, `Atomic<T>`, and atomic reference counts for values shared across threads.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.

Cycles: the identity-cycles node decides how reference cycles through `Shared`, channels and tasks are handled. Implement its answer here.
