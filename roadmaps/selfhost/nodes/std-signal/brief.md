---
id: std-signal
title: "std/signal: SIGINT and SIGTERM for clean shutdown"
status: active
priority: should
type: feature
story: batch-24
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/std-signal/**
  - std/src/signal.nra
  - testdata/signal*
shared:
  - selfhost/runtime/*
  - selfhost/src/runtime*.nra
  - selfhost/src/std/stdtext.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# std/signal: SIGINT and SIGTERM for clean shutdown

## Goal

A program can wait for SIGINT or SIGTERM (Ctrl-C and Ctrl-Break on Windows) as an async event, so a server stops accepting work and exits cleanly instead of dying. The default action stays in place until a program asks for a signal.

## Context

- Source: std-gap-survey rank 26; `inbox.md` 2026-11-20 (Crystal comparison); `GOAL.md` Decisions, 2026-11-21 (signals stay in std).
- The async loop uses kqueue, epoll and `WSAPoll` (SPEC section on `std/async`). Deliver the signal through the loop: a self-pipe or `signalfd` on Linux, `EVFILT_SIGNAL` on macOS, `SetConsoleCtrlHandler` on Windows. The handler itself only writes a byte.
- On `Wasi-Wasm32` the functions are E0991, like `std/net` (SPEC, wasi target rules).

## Acceptance

- [ ] A sample waits for SIGTERM, prints a line and exits 0. The test harness sends the signal with `kill` and checks the output.
- [ ] A second signal during shutdown still reaches the program (the handler stays installed).
- [ ] A wasi build that calls the module is E0991.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

M: runtime C per platform and one loop source.
