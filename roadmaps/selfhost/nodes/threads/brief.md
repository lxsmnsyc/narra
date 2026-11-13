---
id: threads
title: "Threads runtime and std/thread: thread mode, tagged counts, per-thread heaps"
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
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent: 460438
owns:
  - roadmaps/selfhost/nodes/threads/**
  - std/src/thread.nra
  - selfhost/src/stdthread.nra
  - selfhost/runtime/std_thread.c
  - selfhost/src/runtimestdthread.nra
  - testdata/threads_*
shared:
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimecore.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emitdynamic.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/emitconstvalues.nra
  - selfhost/src/driver.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdtext.nra
  - selfhost/runtime/std_async_loop.c
  - compiler/README.md
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: threads-design
    kind: hard
    why: D111 fixes the count encoding, the heap design and the std/thread functions this node builds
  - on: threads-process
    kind: hard
    why: Process.run and the test runner must be safe from several threads before a program can start one
---
# Threads runtime and std/thread: thread mode, tagged counts, per-thread heaps

## Goal

Build D111 items 2, 3 and 5, the panic and constant changes, and `std/thread` (`Thread.scope`, `Thread.start`, `Thread.join`). A program that imports `std/thread` is built in thread mode (`#define NARRA_THREADS 1`); every other program emits the same C as before. The compile-time rules (Sendable, Scope escape, thread function form) are threads-check's.

## Context

- Read D111 and the threads-design outcome first; they override this brief where they differ. Measurements: `nodes/threads-design/log.md`.
- **Counts.** Every count operation in the emitter (`emitdecls.nra` retain, release and unique helpers per type; `emitclosures.nra` environments; `emitdynamic.nra`; heap enum values) and in `runtime.h` (strings, generators, boxes, maps) goes through `NARRA_RC_RETAIN`, `NARRA_RC_RELEASE` and `NARRA_RC_UNSHARE`, which expand to today's code without `NARRA_THREADS`. Immortal moves from -1 to 0 in both modes: static strings (`emitstate.nra`), stack environments (`.h = {-1, NULL}` in `emitclosures.nra`), and the `rc >= 0` guards become `rc != 0`. Check that no code reads -1 as a marker after the change.
- **Share walkers.** One emitted `share` function per counted type, beside its `drop`, emitted only in thread mode: positive counts become negative, stopping at an object already shared. Called on the captures at `Thread.start`; channels and Shared call it later.
- **Heap.** `narra_heap` per thread in thread mode, owner number in the block header above the size class, a lock-free remote-free stack per heap, heaps reused by later threads, main thread on the static heap while no other thread runs (a global running count, plain load). `narra_live` per heap, summed at exit.
- **Constants** (D61): in thread mode the getter checks the ready flag with an acquire load, builds under a lock, and marks the value shared.
- **Panics:** a process lock, flush, code 101, `_exit` while other threads run. `narra_panic_jmp` and `narra_co_panicked` become `_Thread_local`.
- **std/thread** is a source module over `std/internal/thread` (rows in `stdthread.nra`, C in `std_thread.c`, one new case in `runtimetext.nra`). `Thread<T>` and `Scope` are opaque structs over a builtin counted type `SyncSlot<T>` (atomic count, mutex, condition variable, value). POSIX threads with an 8 MiB stack. `driver.nra` adds `-pthread` in thread mode. After a thread function returns, the thread runs its event loop with the D109 exit rule before it ends.
- Proof first: `testdata/threads_scope.nra` fails to compile on the batch base (no `std/thread`).
- Sendable checks do not exist yet: samples must only send Sendable values; threads-check adds the errors.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check: `threads_scope` (the SPEC sample: two threads sum halves of an array); `threads_join_twice`; `threads_shared_array` (eight threads read one captured array of strings; no copy, counts negative, freed once at the end); `threads_write_copies` (a thread changes its captured array in place; the parent's array is unchanged); `threads_alloc_cross` (one thread allocates many small values, another frees them; peak memory stays bounded, checked by running it twice as long and comparing the live count); `threads_constant` (a top-level array constant first read by four threads at once); `threads_tasks` (a thread function starts a task with a timer; join waits for the loop to drain); `threads_deep` (recursion 15000 frames deep on a started thread); `threads_panic` (a panic on a thread exits 101 with its message, in the `panics` or `exits` suite).
- [ ] A program without `std/thread` emits byte-identical C to the batch base except for the -1 to 0 change: check with `narrac emit selfhost` and the bench programs, and record it in the outcome.
- [ ] `narrac emit selfhost` instructions retired within noise of the batch base (`/usr/bin/time -l`, median of 5).
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point.
- [ ] `compiler/README.md` describes thread mode; SPEC "not built yet" notes for this node are removed.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

L: runtime rewrite of every count site (high spread), new heap code with a lock-free stack (high risk), and a new std module; no parser work.

## Notes for the worker

- Do not use threads in `selfhost/src` (seed rule).
- AddressSanitizer builds use `malloc` (`NARRA_PLAIN_MALLOC`); test the heap's remote frees in a release build too.
- macOS `_Thread_local` costs a function call per access outside one function; read the heap pointer once per call.
- The second node to land in `std_thread.c` and `stdthread.nra` (channels, shared) rebases and keeps both sides.
