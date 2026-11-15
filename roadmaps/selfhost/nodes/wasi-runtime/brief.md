---
id: wasi-runtime
title: "Runtime and std on wasm32-wasi"
status: done
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
  - roadmaps/selfhost/nodes/wasi-runtime/**
  - testdata/wasi_*
shared:
  - selfhost/runtime/*
  - selfhost/src/runtime*.nra
  - std/src/*.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/driver.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: target-table
    kind: behavior
    via: targets
---
# Runtime and std on wasm32-wasi

## Goal

Programs built with `--target Wasi-Wasm32` run under `wasmtime`. The std modules that WASI preview 1 cannot back are left out on that target with `@target`, so using them is E0991 at compile time, not a link error.

## Context

Probe 2026-11-14 (zig 0.17.1, wasmtime 50.0.1):

- A small program builds and runs with `zig cc -target wasm32-wasi -mexception-handling -mllvm -wasm-enable-sjlj ... -lsetjmp` and `wasmtime -W exceptions=y`. Without those flags `setjmp.h` fails: the panic handler (`runtime.h`, `narra_panic_jmp`) needs setjmp. Those flags go into `targetCompilerArguments` for the wasm row. They are not enough: the row also needs `-mllvm -wasm-use-legacy-eh=false`, or wasmtime 50 refuses any module that links `setjmp`, and the emitted C must define the `__c_longjmp` tag. See D122 (item 5 and the probe table). The `targets` contract lists neither; this node carries them.
- `seed/narrac.c` for wasm fails on `signal.h`, `sys/wait.h`, `pipe`, `kill`, `sigaction`: `std_process.c`, `std_test.c` and `test_runner.c`. `std_async_io.c` uses sockets and `std_async_loop.c` uses kqueue or epoll; `std_thread.c` uses pthreads (wasi threads are not in preview 1).

## Work

1. Mark the std functions of `std/process`, `std/thread`, the socket half of `std/async`, and `std/terminal` sizes `@target(os = ...)` without `Os.Wasi`, so wasi builds never reach their C. Keep one signature for every target (SPEC section 12).
2. The async loop on wasi: timers and file reads through `poll_oneoff`, or a loop with no I/O sources. Decide in the node and say which in SPEC.
3. `narrac test` on wasi: the test runner forks per test. Either run tests in one process (a panic ends the run) or say that `test --target Wasi-Wasm32` is not supported (new error). Prefer the error unless the one-process form is small.
4. A sample set: every `testdata/*.nra` that uses no left-out module builds and runs under wasmtime with the same output. Record which samples are skipped and why in the node's outcome.
5. `Env.args` on wasi reads `args_sizes_get` and `args_get`. Today it reads `/proc/self/cmdline` everywhere but macOS, so every wasi program sees no arguments.
6. `Fs.writeText` to `/dev/stdout` and `/dev/stderr` writes descriptors 1 and 2 on wasi, because WASI has no `/dev`. narrac prints all its output this way.

Items 5 and 6 were added from D123 (item 7) and the js-interop-design spike. The spike's patches are in that node's log; read them and write the fix to fit this runtime.

## Acceptance

- [ ] The sample set passes under wasmtime, run by node cross-targets' script (this node runs it by hand before that script exists).
- [ ] A wasi build that calls `Process.run` is E0991.
- [ ] A wasi build of a sample that prints its arguments, and writes through `/dev/stdout`, matches its golden under wasmtime.
- [ ] `sh scripts/test.sh` passes on the host; the native runtime is unchanged (debug copy check exit 0).

## Sizing

M: flags, `@target` marks on std, one loop decision. Split if the async loop needs a real `poll_oneoff` backend.
