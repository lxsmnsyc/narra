---
id: async-thread-functions
title: "A thread function may be async"
status: done
priority: could
type: feature
story: batch-22
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/async-thread-functions/**
shared:
  - selfhost/**
  - std/src/thread.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - docs/guide/15-threads-and-shared-state.md
  - docs/guide/14-async.md
  - docs/reference/**
reserves: []
provides: []
depends_on:
  - on: std-async-fold
    kind: hard
    why: sleep and net became async only there
---
# A thread function may be async

## Goal

std-async-fold made `Time.sleep` and `std/net` async only (D125). A thread function cannot be `async` (E1192, SPEC section 8 "Thread functions"), so code running on a thread can no longer sleep or use the network: `testdata/channel_full_blocks.nra` had to replace a sleep with a busy wait on `Time.now()`. Each thread already has its own event loop (SPEC: "Tasks on a thread").

Let a thread function be `async` (an async arrow at the call or the name of a top-level async function): the thread runs the function's task on its own loop until it finishes, then cancels and drains what is left as today, and `join` gives the awaited result. Captures and the result stay `Sendable` (E1190); an `Async<T>` still never crosses threads. Remove E1192 (no migration error). `Parallel.map` callbacks and `Shared.update` callbacks stay plain (decide `Parallel.map` with a reason; the lock rule for `Shared.update` stays). Then:

- `testdata/channel_full_blocks.nra` goes back to a real sleep in an async thread function.
- SPEC section 8, DECISIONS (amend D125 or add a line), chapter 15 ("Threads or async tasks" and the thread function rules) and chapter 14 (one sentence) teach it; regenerate the std reference and error index if they change.

## Acceptance

- [ ] An async thread function that sleeps and uses `std/net` on loopback runs, under the sanitizer and leak check; a thread function that awaits returns its awaited value through `join`.
- [ ] Captures and results that are not `Sendable` are still rejected; the busy wait is gone.
- [ ] `sh scripts/test.sh` passes; `pnpm --dir website build` passes.
