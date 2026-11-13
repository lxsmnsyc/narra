---
id: async-generators
title: "Async generators and for await"
status: todo
priority: must
type: feature
story:
early: false
outline: true
milestone: p1-syntax
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: async-core
    kind: hard
    why: builds on Async<T>, the event loop and the async state machine (D93)
---

# Async generators and for await

## Goal

`async function*` returning `AsyncGenerator<T>`, consumed with `for await (const chunk of stream)`, as in TypeScript (D95). Needed for HTTP bodies, server-sent events and file streams, so it lands before the `std/async` I/O modules return streams. `ensures` applies to each yielded chunk.

## Notes

Outline. Decide when refined: cancellation of a stream mid-iteration (D93 cooperative cancel), backpressure, and the time operators the inbox lists (`debounce`, `throttle`, `merge`).
