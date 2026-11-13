---
id: threads-shared
title: "std/shared and std/atomic: Shared<T> and Atomic<T>"
status: todo
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
spent:
owns:
  - roadmaps/selfhost/nodes/threads-shared/**
  - std/src/shared.nra
  - std/src/atomic.nra
  - testdata/shared_*
  - testdata/atomic_*
  - testdata/errors/shared_*
  - testdata/errors/atomic_*
  - testdata/leaks/shared_*
shared:
  - selfhost/src/stdthread.nra
  - selfhost/runtime/std_thread.c
  - selfhost/src/runtimestdthread.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/threadcheck.nra
  - selfhost/src/diagnostics.nra
  - compiler/README.md
  - SPEC.md
reserves: [diag E1410-E1419]
provides: []
depends_on:
  - on: threads-design
    kind: hard
    why: D111 item 4 fixes the Shared and Atomic functions and codes
  - on: threads-check
    kind: hard
    why: the value type must be Sendable, through the check that node builds
---
# std/shared and std/atomic: Shared<T> and Atomic<T>

## Goal

Build `Shared<T>` and `Atomic<T>` of D111 item 4: `Shared.create`, `get`, `set`, `update`, `read`, and `Atomic.create`, `load`, `store`, `exchange`, `add`, `compareExchange`, as SPEC section 8 "Shared state" describes them.

## Context

- Read D111 and the outcomes of threads and threads-check.
- Both are source modules over `std/internal/thread`; the types are opaque structs over `SyncSlot<T>`. The names and callback forms follow `std/cell` (`std/src/cell.nra`, D90).
- `Shared.update` holds the lock, takes the value out of the slot, makes a top count of -1 into 1 so the callback can change the top object in place, stores the result and marks it shared. `set` and `create` mark the value shared. `get` returns a copy that holds one more (atomic) reference.
- `Atomic<T>` for `bool` and integer types up to 64 bits, C11 atomics, sequentially consistent. `add` wraps (D92) and gives the old value.
- The Shared SPEC sample is in `selfhost/tests/spec-skip.txt`; remove its line when it checks.
- Call the threads-check helper for E1190 on `T` of `Shared`, and add `Shared` and `Atomic` to its E1195 list.
- node threads-channels edits `stdthread.nra` and `std_thread.c` at the same time; the second to land rebases and keeps both sides.

## Codes (E1410-E1419)

- E1410 an `async` callback to `Shared.update` or `Shared.read`.
- E1411 an `Atomic<T>` whose `T` is not `bool` or an integer type (in a type or at `Atomic.create`).
- E1412-E1419 free.

## Acceptance

- [ ] Proof first: `testdata/shared_counter.nra` fails to compile on this node's base.
- [ ] Samples with checked output, clean under AddressSanitizer and the leak check: `shared_counter` (eight threads, 10000 `update` calls each, total checked); `shared_array_in_place` (`update` pushes to an array held only by the `Shared`; the debug copy check reports no copy); `shared_read_parts` (`read` returns a part of the value, used after another thread's `set`); `atomic_counter` (eight threads `add`; the result is exact); `atomic_compare_exchange` (a spin flag); `atomic_wrap` (`Atomic<u8>` wraps at 255).
- [ ] Leak test: `testdata/leaks/shared_cycle.nra`, a `Shared` whose value holds itself, reported by the leak check (D80).
- [ ] Error tests: `shared_async_callback` (E1410), `atomic_float` (E1411), `shared_cell_value` (E1190), `shared_equal` (E1195).
- [ ] SPEC "not built yet" note for Shared state removed; `compiler/README.md` updated.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

M: two source modules, about 150 lines of C, two small checks.

## Notes for the worker

- Nested `Shared` calls in different orders can deadlock; nothing reports them (SPEC). Do not add a runtime check.
