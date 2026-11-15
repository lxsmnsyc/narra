---
id: windows-process-async
title: "std/process and the async loop on Windows x64"
status: todo
priority: could
type: feature
story: batch-16
early: false
outline: false
milestone: p5-distribution
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
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/windows-process-async/**
  - testdata/windows_process*
  - testdata/windows_async*
shared:
  - selfhost/runtime/std_process.c
  - selfhost/runtime/std_async_loop.c
  - selfhost/runtime/std_async_io.c
  - selfhost/runtime/std_test.c
  - selfhost/runtime/test_runner.c
  - selfhost/src/runtime*.nra
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: windows-runtime
    kind: hard
    why: builds on its Windows test setup and the runtime branches it adds; both edit the same runtime C files
---
# std/process and the async loop on Windows x64

## Goal

`std/process`, `narrac test` and `std/async` work on Windows, so the compiler itself (`seed/narrac.c`) builds and runs there.

## Context

The POSIX code: `std_process.c` (fork, pipe, poll, waitpid, signals), `std_test.c` and `test_runner.c` (fork per test, signals), `std_async_loop.c` (kqueue on macOS, epoll on Linux), `std_async_io.c` (BSD sockets). On Windows: `CreateProcessW` with pipes and a thread per stream instead of poll, a job object for timeouts and kill, a child process per test instead of fork (re-run the binary with the test name), and a loop on `WSAPoll` (simplest) or IOCP.

## Acceptance

- [ ] `seed/narrac.c` built for Windows compiles the samples and the output matches the host's (run through cross-targets' script).
- [ ] The process and async samples pass on Windows; any that cannot are listed in SPEC as known gaps.
- [ ] `sh scripts/test.sh` passes on the host.

## Sizing

M if the loop uses WSAPoll; split off an IOCP loop as its own node if WSAPoll is not enough.
