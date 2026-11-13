---
id: threads-process
title: "Thread-safe signal forwarding in Process.run, the test runner and Test.panics"
status: done
priority: must
type: fix
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent: 92496
owns:
  - roadmaps/selfhost/nodes/threads-process/**
shared:
  - selfhost/runtime/std_process.c
  - selfhost/src/runtimestdprocess.nra
  - selfhost/runtime/test_runner.c
  - selfhost/src/runtimetestrunner.nra
  - selfhost/runtime/std_test.c
  - selfhost/src/runtimestdtest.nra
reserves: []
provides: []
depends_on:
  - on: threads-design
    kind: hard
    why: D111 item 6 fixes the forwarding design this node builds
---
# Thread-safe signal forwarding in Process.run, the test runner and Test.panics

## Goal

Build D111 item 6. `Process.run`, the test runner and `Test.panics` each keep one process-global child pid and install the SIGTERM, SIGINT and SIGHUP forwarding handlers per call with `sigprocmask`. Make them safe for calls from several threads at once, before node threads lets a program start a thread.

## Context

- Source: D111 item 6; the process-signals outcome ("the handler state is process-global").
- Per file: a mutex-guarded count of active calls; the first call saves and installs the handlers, the last restores them. Child pids live in a fixed array (a slot per active call, for example 64, with a lock-free claim) that the handler walks, so a signal reaches every running child. `pthread_sigmask` replaces `sigprocmask`. The child side of the fork restores the saved handlers and mask as today.
- The generated `runtime*.nra` files follow from `sh scripts/test.sh --suite generated --update`. Check the generated file names with `ls selfhost/src/runtime*`.
- There is no thread API yet, so the behavior with one caller must stay exactly as process-signals built it; the existing process-signals tests are the proof. Add `-pthread` only if a file needs it to link (mutexes need no flag on macOS or glibc 2.34+).

## Acceptance

- [ ] The process-signals tests still pass: SIGTERM, SIGINT and SIGHUP reach the child of `Process.run`, `narrac run` and `narrac test`; the caller survives; `Process.run` reports 128 plus the signal; an ignored signal stays ignored.
- [ ] A C-level check in the outcome (a scratch C program that includes the runtime text and calls `narra_proc_run` from two pthreads while a signal arrives) shows both children get the signal and the handlers are restored after both calls. It is evidence, not a committed test, since threads have no Narra API yet; node threads adds a Narra sample.
- [ ] Fast check passes: `sh scripts/test.sh`.

## Sizing

S: three C files with the same pattern, no compiler change.

## Notes for the worker

- Editing `selfhost/runtime/` text needs two seed updates at the end of the batch; do not regenerate the seed yourself.
- A signal handler may only call async-signal-safe functions: read the pid array with plain loads and call `kill`, no mutex in the handler.
