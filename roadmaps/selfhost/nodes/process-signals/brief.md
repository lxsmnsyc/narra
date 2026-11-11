---
id: process-signals
title: "narrac run and test forward termination signals"
status: active
priority: should
type: fix
story: batch-9
early: false
outline: false
milestone: 
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/process-signals/**
shared:
  - selfhost/runtime/std_process.c
  - selfhost/src/runtimestdprocess.nra
  - selfhost/src/driver.nra
  - testdata/**
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# narrac run and test forward termination signals

## Goal

Inbox, from batch-6: `narrac run` and `narrac test` leave the program running when they are killed. `Process.run` in `selfhost/runtime/std_process.c` forks and waits, with no process group and no signal forwarding, so a timeout that kills `narrac` orphans a hung program (a probe left 15 programs spinning for 6 hours). Forward SIGTERM, SIGINT and SIGHUP to the child, or start it in its own process group and kill the group; pick the one that keeps terminal Ctrl-C working for `narrac run`. Write the choice in SPEC (std/process) in one line.

## Acceptance

- Fix proof: on main, `timeout` (or `kill`) on `narrac run` of a looping program leaves the program alive; on the branch it is gone. Record both in `log.md`, and add a test the runner can run without leaving processes behind.
- Works on macOS and Linux (`sh scripts/linux-check.sh`).
- The runtime is edited in `selfhost/runtime/`; regenerate the generated text with `sh scripts/test.sh --suite generated --update`. Do not regenerate the seed.
- `sh scripts/test.sh` passes.
