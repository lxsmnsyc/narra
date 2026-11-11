---
id: process-exit
title: "A std function that ends the process with an exit code"
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
  - roadmaps/selfhost/nodes/process-exit/**
  - selfhost/src/stdprocess.nra
shared:
  - selfhost/runtime/std_process.c
  - selfhost/src/runtimestdprocess.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/builtins.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/diagnostics.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - docs/guide/**
reserves: []
provides: []
depends_on: []
---

# A std function that ends the process with an exit code

## Goal

Inbox, from cli-skip-teardown: `skipTeardown` in `selfhost/src/emitter.nra` keeps the Checker and Package alive through a `Cell` that holds itself, so the process exit frees them instead of the compiler. Replace that stopgap with a std function that exits with a code, for example `Process.exit(code: int): never uses io` in `std/process`. Decide and document in SPEC:

- the effect it charges (`io` or a new one) and its result type (it does not return);
- that it skips releasing live values, so the leak check (`NARRA_CHECK_LEAKS`) must not count them, or the function releases nothing and the leak check is told to skip;
- whether pending stdout is flushed (it must be).

The compiler itself does not use the function in this node (seed rule); seed-switch does that after the mid-batch seed update.

## Acceptance

- A sample program exits with the code it passes, with output flushed, clean under AddressSanitizer and the leak check.
- SPEC and the guide's std chapter list the function.
- `sh scripts/test.sh` passes.
