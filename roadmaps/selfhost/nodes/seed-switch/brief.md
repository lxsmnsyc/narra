---
id: seed-switch
title: "The compiler uses Process.exit and the host-target builtin"
status: todo
priority: should
type: refactor
story: batch-9
early: false
outline: false
milestone: 
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/seed-switch/**
  - selfhost/src/emitter.nra
  - selfhost/src/buildtarget.nra
shared:
  - selfhost/src/driver.nra
  - selfhost/src/emit.nra
  - seed/**
reserves: []
provides: []
depends_on:
  - on: process-exit
    kind: hard
    why: the compiler can call the function only after a seed update that knows it
  - on: host-target-builtin
    kind: hard
    why: the compiler can call the builtin only after a seed update that knows it
  - on: process-signals
    kind: hard
    why: its runtime change rides the same mid-batch seed update
  - on: test-runner-fixes
    kind: hard
    why: its std text change rides the same mid-batch seed update
  - on: ptr-cast
    kind: hard
    why: its std functions ride the same mid-batch seed update
---

# The compiler uses Process.exit and the host-target builtin

## Goal

Run by the orchestrator after every other child has landed: `sh scripts/seed-update.sh`, then

- replace `skipTeardown` and the `Retained` cycle in `selfhost/src/emitter.nra` (callers in `emitter.nra` and `driver.nra`) with the std exit function from process-exit, keeping the leak check's full release under `NARRA_CHECK_LEAKS`. If it is safe, also skip the Emitter release inside `emitProgram` (`emit.nra`), which the cli-skip-teardown profile put at part of a 6.3% estimate;
- replace the `uname` call in `selfhost/src/buildtarget.nra` with the host-target builtin, keeping `NARRA_HOST` (or the existing override) for tests;

then `sh scripts/seed-update.sh` again.

## Acceptance

- `sh selfhost/bootstrap.sh` reports the seed current; `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass.
- `narrac emit selfhost` instructions measured before and after the switch, recorded in `outcome.md`.
