---
id: seed-switch-cleanup
title: "narrac run exits without a release on a failed start, and the host target matches by value"
status: done
priority: should
type: refactor
story: batch-10
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
added: 2026-11-12
spent: 118000
owns:
  - roadmaps/selfhost/nodes/seed-switch-cleanup/**
  - selfhost/src/buildtarget.nra
shared:
  - selfhost/src/driver.nra
reserves: []
provides: []
depends_on: []
---

# narrac run exits without a release on a failed start, and the host target matches by value

## Goal

Two findings from the seed-switch review (batch-9):

1. In `selfhost/src/driver.nra` (the `run` path, near line 786), when `startProgram` fails after a successful build, the function returns 1 and the compiler does a full release. The other exits on that path go through `exitWithoutRelease`. Use it here too, keeping the leak check's full release under `NARRA_CHECK_LEAKS` as `exitWithoutRelease` already does. Check the other early `return 1` lines in the same function for the same gap.
2. `selfhost/src/buildtarget.nra` (near line 70) turns `Build.hostTarget()` into text (`describeOs`, `describeArch`) and compares it with each row's `os-arch` text. Compare the `Os` and `Arch` values with the rows directly, so the text helpers are needed only where text is printed. Keep the `NARRA_HOST` override, which names a row by text.

## Acceptance

- `narrac run` on a program that builds but cannot start still prints the error and exits 1.
- `buildtarget.nra` has no text comparison of the host target.
- `sh scripts/test.sh` passes.
