---
id: std-async-io
title: "std/async/net and std/async/file"
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
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent: 218448
owns:
  - std/src/async__net.nra
  - std/src/async__file.nra
  - selfhost/src/stdasyncio.nra
  - selfhost/runtime/std_async_io.c
  - selfhost/src/runtimestdasyncio.nra
  - testdata/async_io_*
  - testdata/errors/async_io_*
  - roadmaps/selfhost/nodes/std-async-io/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdtext.nra
  - selfhost/runtime/std_async.c
  - selfhost/runtime/std_async_loop.c
  - selfhost/src/runtimestdasync.nra
  - selfhost/src/runtimestdasyncloop.nra
  - selfhost/src/types.nra
  - selfhost/src/resolve.nra
  - selfhost/src/check.nra
  - selfhost/src/exprs.nra
  - selfhost/src/cnames.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/traits.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/sharedresults.nra
  - selfhost/src/stdcopy.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/spec-skip.txt
  - compiler/README.md
reserves: [diag E1170-E1189]
provides: []
depends_on:
  - on: event-loop
    kind: hard
    why: the suspending I/O registers with the loop's poller
  - on: async-api
    kind: hard
    why: the cancelled-read test needs Async.cancel, and both change the cancel path in the async runtime
---
# std/async/net and std/async/file

## Goal

The suspending I/O modules of SPEC section 8, as D109 item 8 sets them: `std/async/net` (`listen`, `localPort`, `accept`, `connect`, `read`, `write`) and `std/async/file` (`readText`, `writeText`, `readBytes`, `writeBytes`, the names and arguments of `std/fs`). Every suspending function returns `Result<T, AsyncError>`, and a cancelled task's next call returns `Err(AsyncError.Cancelled)` without doing the work. `std/async/time` is event-loop's.

## Context

- Read D109 (items 7, 8 and 10) and the event-loop outcome first.
- **Handles.** `Listener` and `Connection` are opaque structs of `std/async/net` over a builtin type `Handle` of `std/internal/asyncio`: a counted runtime object holding a descriptor, whose last release closes it (SPEC section 6). Build it as `Type.TStringBuilder` is built, a builtin type visible only in modules that import its module (13 files name `TStringBuilder`; they are the list to visit). No `==` on it; it is not a contents type, so it is passed by value, never as `mutable`. Under `NARRA_CHECK_LEAKS` an open handle counts as a live allocation.
- **Net.** Non-blocking sockets registered with the loop's poller (event-loop's registration functions). `host` is a numeric IPv4 or IPv6 address, no name lookup. `listen(host, 0)` binds a free port and `localPort(listener)` reads it, so parallel test shards do not collide. `read` gives empty bytes at the end of the stream.
- **Files.** kqueue and epoll do not wait on regular files. Each call does the blocking read or write, then puts the task at the back of the ready queue once, so other ready tasks run. A cancelled task's call returns `Err(AsyncError.Cancelled)` before it touches the file.
- Nested std modules are flat files: `std/async/net` is `std/src/async__net.nra`. Its primitives are the builtin module `std/internal/asyncio` (rows in `stdasyncio.nra`, registered in `stdlib.nra`), whose C is `selfhost/runtime/std_async_io.c`; add the `runtimetext.nra` case for `std/internal/asyncio`, then `sh scripts/test.sh --suite generated --update`.
- The SPEC section 8 sample (an echo server) checks once this node lands: remove its line from `selfhost/tests/spec-skip.txt` (the `spec` suite fails while a skipped sample passes).
- Size L, not M: the `Handle` type adds cases to the checker and emitter beside the runtime and two modules. If the worker finds it larger than L, the orchestrator splits off `std/async/file`.

## Codes (E1170-E1189)

None expected; a closed or failed handle is a runtime `Err(AsyncError.Failed(message))`. Use the block only if a compile-time rule turns up (for example a `Handle` used outside std), and record it in the outcome.

## Acceptance

- [ ] Samples, output checked, clean under AddressSanitizer and the leak check, on macOS and in the Linux check: an echo server and client on loopback in one program (port 0); an async file write then read; a cancelled read that returns `Err(AsyncError.Cancelled)`; a connection closed when its last copy is dropped (the peer reads end of stream); a read from a peer that has closed.
- [ ] The SPEC section 8 sample passes the `spec` suite, and its skip line is gone.
- [ ] `compiler/README.md` lists the first-build limits: file calls block the loop for their length, no name lookup.
