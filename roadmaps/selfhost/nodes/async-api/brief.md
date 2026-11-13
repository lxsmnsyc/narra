---
id: async-api
title: "Async.then, detach, peek, cancel, all and race"
status: active
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
  - selfhost/runtime/std_async.c
  - selfhost/src/runtimestdasync.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/diagnostics.nra
reserves: [diag E1150-E1169]
provides: []
depends_on:
  - on: async-core
    kind: hard
    why: the functions act on the handle layout async-core builds
---
# Async.then, detach, peek, cancel, all and race

## Goal

Build the `std/async` functions of D93 with the signatures of D109: `Async.then(handle, escaping callback)`, `Async.peek(handle): Option<T>`, `Async.cancel(handle)`, `Async.all(handles: Array<Async<T>>): Async<Array<T>>` and `Async.race(handles: Array<Async<T>>): Async<Option<T>>` (`None` for an empty array). async-core has built `Async.detach`.

## Context

- Read D109 (items 2, 8 and the `std/async` functions bullet) and the async-core outcome first.
- The functions are Narra source in `std/src/async.nra` over primitives of `std/internal/async` (rows in `stdasync.nra`, C in `selfhost/runtime/std_async.c`). Edit the sources, then `sh scripts/test.sh --suite generated --update`.
- `then` stores the callback in the task header; `narra_task_finish` runs the callbacks unless the task was cancelled, then drops them. `cancel` sets the flag and drops the callbacks; the I/O side of cancel (waking a task that waits on a timer or a descriptor) is event-loop's and std-async-io's. `peek` gives `None` from the cancel on. `all` and `race` are tasks that await their inputs; neither cancels them.
- `Arrays.map(items, async (item) => ...)` already gives `Array<Async<T>>` (async-core); `all` takes that array.
- event-loop changes `std_async.c` and `std/src/async.nra` at the same time; the second to land rebases and keeps both sides.

## Codes (E1150-E1169)

None expected: a misuse of these functions is an ordinary type error. Use the block only if a compile-time rule turns up, and record it in the outcome.

## Acceptance

- [ ] Samples for each function, output checked, clean under AddressSanitizer and the leak check: `then` on a pending and on a finished handle; a cancelled handle whose `then` callbacks do not run and whose callbacks are freed; `peek` before finishing, after finishing and after a cancel; `await` on a cancelled handle still gives the body's value; `cancel` on a finished handle does nothing; `all` keeps the input order when the tasks finish out of order; `race` gives the first value, and `None` for an empty array.
- [ ] A `then` callback that captures the handle it is registered on does not leak after it runs (the runtime drops it).
