---
id: event-loop
title: "The event loop: kqueue, epoll, timers and exit"
status: done
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
spent: 447582
owns:
  - testdata/async_loop_*
  - testdata/async_time_*
  - testdata/async_idle_*
  - testdata/async_exit_*
  - std/src/async__time.nra
  - selfhost/runtime/std_async_loop.c
  - selfhost/src/runtimestdasyncloop.nra
  - roadmaps/selfhost/nodes/event-loop/**
shared:
  - selfhost/src/stdasync.nra
  - std/src/async.nra
  - selfhost/runtime/std_async.c
  - selfhost/src/runtimestdasync.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/diagnostics.nra
  - scripts/linux-check.sh
reserves: [diag E1130-E1149]
provides: []
depends_on:
  - on: async-core
    kind: hard
    why: replaces the ready-queue loop async-core adds
---
# The event loop: kqueue, epoll, timers and exit

## Goal

Give the D109 loop real waiting (D109 item 7): `kqueue` on macOS, `epoll` on Linux, a timer heap, `std/async/time` with `sleep(duration: Duration): Result<Duration, AsyncError>`, and `await Async.idle()` (D95). The program exits when the entry function finishes; tasks still running are cancelled, and a cancelled task's pending sleep returns `Err(AsyncError.Cancelled)` (D109 items 6 and 8).

## Context

- Read D109 (items 6, 7, 8 and 10) and the async-core outcome first. There is no `Async.sleep`: sleep lives in `std/async/time`, with the name and argument of `Time.sleep`.
- The loop is `selfhost/runtime/std_async_loop.c`, which async-core created as a ready queue. Each turn runs the ready queue in order, then waits in the poller until the nearest timer deadline. The timers are a binary min-heap ordered by deadline, then by creation order. `kqueue` under `__APPLE__`, `epoll` under `__linux__`, `#error` elsewhere. The poller registration functions take a descriptor and a task, so std-async-io can add sockets without changing the loop.
- `std/async/time` is a new source module, `std/src/async__time.nra` (nested std modules are flat files: `std/ext/array` is `ext__array.nra`), registered in `stdlib.nra` with its primitives in `stdasync.nra`. `Async.idle` goes in `std/src/async.nra`.
- Generated files: edit `std_async_loop.c` and the `.nra` sources, then `sh scripts/test.sh --suite generated --update` writes `runtimestdasyncloop.nra` and `stdtext.nra`. async-core's `runtimetext.nra` case already returns the loop text.
- async-api changes `std_async.c` and `std/src/async.nra` at the same time; the second to land rebases and keeps both sides.
- `sh scripts/linux-check.sh` runs the epoll path; run it before landing.

## Codes (E1130-E1149)

None expected: `Async.idle` and `std/async/time` are ordinary calls. Use the block only if a compile-time rule turns up, and record it in the outcome.

- async-core built `Async.idle()` with a looser rule (it finishes once no task is ready). Make it follow D109 item 7: it finishes when no other task of the loop is unfinished, including tasks started while it waits.

## Acceptance

- [ ] Samples, output checked and clean under AddressSanitizer and the leak check on macOS and in the Linux check: several sleeping tasks finish in deadline order, and two equal deadlines in the order they were set; `Async.idle` waits for tasks started while it waits; a sleeping task cancelled with `Async.cancel` gets `Err(AsyncError.Cancelled)` (once async-api has landed; before that, through the exit sequence); tasks still sleeping at exit are cancelled, run to their return and leak nothing; a detached task that never finishes does not keep the program running after the entry returns.
- [ ] The exit code is the entry's value whatever the other tasks do.
