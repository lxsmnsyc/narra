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

Build D93: `Async<T>` as an eager shared handle, async functions and `async { ... }` blocks, `await`, `Async.then`, `detach`, `peek`, `cancel`, `all` and `race`, one implicit event loop per thread with the entry function as its first task, and `std/async/*` I/O. Rewrite SPEC.md section 8 to match.

## Notes

Outline. Split into children when refined (rough plan, each M):
- async-transform: async functions and blocks compile to the generator state machine (D19, D79); `Async<T>` type, unused-handle error, effects charged at creation.
- event-loop: the runtime loop (kqueue on macOS, epoll on Linux), the entry function as first task, exit when no work is left; gated or stubbed for wasm32-wasi (cross-targets).
- async-api: `then`, `detach`, `peek`, `cancel` (cooperative, `Err(Cancelled)`), `all`, `race`.
- std-async-io: `std/async/net`, `file` and `time` with the blocking modules' names.
Open (D93): the exit code with tasks still running after a failed entry function; loops on several threads.
