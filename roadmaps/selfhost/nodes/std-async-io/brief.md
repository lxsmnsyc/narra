---
id: std-async-io
title: "std/async/net, file and time"
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
  - std/src/async__*.nra
  - selfhost/src/stdasyncio.nra
  - selfhost/runtime/async_io.c
  - selfhost/src/runtimeasyncio.nra
  - testdata/async_io_*
  - roadmaps/selfhost/nodes/std-async-io/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/runtimeasync.nra
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/diagnostics.nra
reserves: []
provides: []
depends_on:
  - on: event-loop
    kind: hard
    why: the suspending I/O registers with the loop's poller
  - on: async-api
    kind: hard
    why: the cancelled-read test needs Async.cancel, and both change the cancel path in the async runtime
---
# std/async/net, file and time

## Goal

The suspending I/O modules of SPEC section 8: `std/async/net`, `std/async/file` and `std/async/time`, with the blocking modules' function names. A cancelled task's next I/O returns `Err(Cancelled)` (D93).

## Context

- Refined by async-design after the D109 gate. Read D109 and the event-loop outcome first.
- Nested std modules are flat files: `std/async/net` is `std/src/async__net.nra` (as `std/ext/array` is `ext__array.nra`), beside async-core's `std/src/async.nra`. Generated files: edit the `.nra` and `.c` sources, add the `runtimetext.nra` case, then `sh scripts/test.sh --suite generated --update`.
- Network samples use the loopback interface and a port the sample picks, so parallel test shards do not collide.

## Acceptance

- [ ] Samples: an echo server and client on loopback in one program, an async file read and write, a cancelled read that returns `Err(Cancelled)`. Clean under AddressSanitizer and the leak check, on macOS and in the Linux check.
