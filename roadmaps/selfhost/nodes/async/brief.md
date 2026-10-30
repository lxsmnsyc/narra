---
id: async
title: Async functions and the event loop
status: todo
priority: later
type: feature
outline: true
milestone: m5-concurrency
target:
from:
size: XL
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
  - on: generators
    kind: hard
    why: async bodies reuse the generator state-machine transform
---

# Async functions and the event loop

## Goal

`async function`, `await`, `Async.run`, `Async.spawn`, `Async.all` and `std/async/*` I/O on an event loop (kqueue or epoll). Split before starting.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
