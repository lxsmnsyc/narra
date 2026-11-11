---
id: host-target-builtin
title: "A builtin that returns the host target from compiler macros"
status: active
priority: should
type: feature
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
  - roadmaps/selfhost/nodes/host-target-builtin/**
shared:
  - selfhost/src/stdbuild.nra
  - selfhost/src/buildtarget.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/builtins.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/runtimetext.nra
  - selfhost/runtime/std_build.c
  - std/src/build.nra
  - testdata/**
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# A builtin that returns the host target from compiler macros

## Goal

Inbox, from build-conditions and batch-7: the host target is found by running `/usr/bin/uname -s -m` when a package imports `std/build` (`selfhost/src/buildtarget.nra`). Checking does I/O, and busybox systems such as Alpine have `uname` only at `/bin/uname`, so the compiler reports "could not determine the host" there. Add a builtin whose C reads the predefined macros (`__APPLE__`, `__linux__`, `__aarch64__`, `__x86_64__`) and returns the target the compiler was built for, mapped to the names in `buildtarget.nra`. An unknown combination returns None.

This node adds the builtin and tests it. The compiler keeps calling `uname` until seed-switch (seed rule).

## Acceptance

- A sample program prints the host target; the test checks it against `uname -s -m` on macOS and in `sh scripts/linux-check.sh`.
- `sh scripts/test.sh` passes.
