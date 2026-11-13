---
id: async-api
title: "Async.then, detach, peek, cancel, all and race"
status: todo
priority: should
type: feature
story: batch-12
early: false
outline: false
milestone: 
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
spent:
owns:
  - testdata/async_api_*
  - testdata/errors/async_api_*
  - roadmaps/selfhost/nodes/async-api/**
shared:
  - selfhost/src/stdasync.nra
  - std/src/async.nra
  - selfhost/src/runtimeasync.nra
  - selfhost/src/diagnostics.nra
reserves: []
provides: []
depends_on:
  - on: async-core
    kind: hard
    why: the functions act on the handle layout async-core builds
---
# Async.then, detach, peek, cancel, all and race

## Goal

Build the `std/async` functions of D93: `Async.then(handle, callback)`, `Async.detach(handle)`, `Async.peek(handle): Option<T>`, `Async.cancel(handle)`, `Async.all(handles)` and `Async.race(handles)`. `Arrays.map(items, async (x) => ...)` gives `Array<Async<T>>`.

## Context

- Refined by async-design after the D109 gate. Read D109 and the async-core outcome first.
- event-loop changes the same runtime text; the second to land rebases and keeps both sides.

## Acceptance

- [ ] Samples for each function, including a cancelled handle whose `then` callbacks do not run and `peek` before and after finishing. Clean under AddressSanitizer and the leak check.
- [ ] Error tests for each code D109 gives this node.
