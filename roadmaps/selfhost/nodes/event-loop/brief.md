---
id: event-loop
title: "The event loop: kqueue, epoll, timers and exit"
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
  - testdata/async_loop_*
  - testdata/async_sleep_*
  - selfhost/runtime/async_loop.c
  - selfhost/src/runtimeasyncloop.nra
  - roadmaps/selfhost/nodes/event-loop/**
shared:
  - selfhost/src/stdasync.nra
  - std/src/async.nra
  - selfhost/src/runtimeasync.nra
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/runtimecore.nra
  - selfhost/src/diagnostics.nra
  - scripts/linux-check.sh
reserves: []
provides: []
depends_on:
  - on: async-core
    kind: hard
    why: replaces the ready-queue loop async-core adds
---
# The event loop: kqueue, epoll, timers and exit

## Goal

Give the D109 loop real waiting: kqueue on macOS, epoll on Linux, a timer heap, `await Async.sleep(milliseconds)` and `await Async.idle()` (D95). The program exits when the entry function finishes; tasks still running are cancelled.

## Context

- Refined by async-design after the D109 gate. Read D109 and the async-core outcome first.
- Generated files: edit `selfhost/runtime/async_loop.c` and `std/src/async.nra`, add the one `runtimetext.nra` case, then `sh scripts/test.sh --suite generated --update` writes `runtimeasyncloop.nra` and `stdtext.nra`.
- `sh scripts/linux-check.sh` runs the epoll path; run it before landing.

## Acceptance

- [ ] Samples: several sleeping tasks finish in timer order; `Async.idle` waits for tasks started while it waits; tasks still running at exit are cancelled and leak nothing. Clean under AddressSanitizer and the leak check on macOS and in the Linux check.
