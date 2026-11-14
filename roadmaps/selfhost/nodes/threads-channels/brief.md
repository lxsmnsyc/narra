---
id: threads-channels
title: "std/channel: bounded channels between threads"
status: done
priority: must
type: feature
story: batch-13
early: false
outline: false
milestone: p1-syntax
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
spent: 312856
owns:
  - roadmaps/selfhost/nodes/threads-channels/**
  - std/src/channel.nra
  - testdata/channel_*
  - testdata/errors/channel_*
  - testdata/leaks/channel_*
shared:
  - selfhost/src/stdthread.nra
  - selfhost/runtime/std_thread.c
  - selfhost/src/runtimestdthread.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/threadcheck.nra
  - selfhost/src/diagnostics.nra
  - compiler/README.md
  - SPEC.md
reserves: [diag E1400-E1409]
provides: []
depends_on:
  - on: threads-design
    kind: hard
    why: D111 item 4 fixes the channel functions and their results
  - on: threads-check
    kind: hard
    why: the element type must be Sendable, through the check that node builds
---
# std/channel: bounded channels between threads

## Goal

Build the channels of D111 item 4: `Channel.create<T>(capacity)`, `send`, `receive`, `values` and `close`, as SPEC section 8 "Channels" describes them.

## Context

- Read D111 and the outcomes of threads and threads-check.
- `std/channel` is a source module; `create<T>` is a generic function, so `Channel.create<int>(16)` uses ordinary call type arguments (no E0326 exception). `Sender<T>` and `Receiver<T>` are opaque structs; each holds its own counted end object over the shared queue, so the last release of every sender closes the channel and the last release of every receiver makes `send` return `false`.
- The queue is a ring buffer of `capacity` slots (below 1 means 1) under the `SyncSlot` mutex, with two condition variables. `send` marks the value shared (the share walker of node threads) before it enters the queue.
- `values` is a `Generator<T> uses io`. The remaining SPEC sample (section 8 "Channels") is in `selfhost/tests/spec-skip.txt`; remove its line when it checks.
- Call the threads-check helper for E1190 on `T`, and add `Sender` and `Receiver` to its E1195 list.
- node threads-shared edits `stdthread.nra` and `std_thread.c` at the same time; the second to land rebases and keeps both sides.

## Codes (E1400-E1409)

None expected: misuses are type errors and E1190 covers the element type. Use the block only if a compile-time rule turns up, and record it in the outcome.

## SyncSlot

- Orchestrator, 2026-11-13: the threads node built `Thread<T>` and `Scope` as boxes in Narra source and left out the builtin `SyncSlot<T>` (a value next to a mutex and condition variable). This node builds it, owns the type, cnames, traits, generics and dump changes it needs, and records its API in the outcome for threads-shared.

## Acceptance

- [ ] Proof first: `testdata/channel_basic.nra` fails to compile on this node's base.
- [ ] Samples with checked output, clean under AddressSanitizer and the leak check: `channel_basic` (the SPEC sample); `channel_full_blocks` (capacity 1, a slow receiver, order kept); `channel_close_drop` (the last sender dropped without `close` ends `values`); `channel_receivers_gone` (`send` returns `false` and the value is freed); `channel_many_receivers` (four threads share a `Receiver`, every value taken once); `channel_reply` (a `Sender<Reply>` inside a request, as the SPEC paragraph says); `channel_capacity_zero` (behaves as 1).
- [ ] Leak test: `testdata/leaks/channel_cycle.nra`, a sender sent through its own channel, reported by the leak check (D80).
- [ ] Error tests: `channel_cell_element` (`Channel.create<Cell<int>>`, E1190), `channel_equal` (E1195).
- [ ] SPEC "not built yet" note for channels removed; `compiler/README.md` updated.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

M: one source module, about 150 lines of C, no new checker rule.

## Notes for the worker

- A blocking call in async code blocks the thread's loop (D111 item 5); there is no async receive in this node.
