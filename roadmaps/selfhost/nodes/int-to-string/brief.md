---
id: int-to-string
title: "Format integers without snprintf"
status: done
priority: should
type: fix
story: batch-4
early: false
outline: false
milestone: m6-performance
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
added: 2026-11-10
spent: 60586
owns:
  - roadmaps/selfhost/nodes/int-to-string/**
shared:
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimecore.nra
reserves: []
provides: []
depends_on:
  - on: fast-alloc
    kind: hard
    why: both edit selfhost/runtime/runtime.h; the earlier node lands first
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# Format integers without snprintf

## Goal

`lstr_from_int` and `lstr_from_uint` in `selfhost/runtime/runtime.h` write the digits into a stack buffer themselves instead of calling `snprintf`. Regenerate `selfhost/src/runtimecore.nra` with `--suite generated --update`.

## Context

- Evidence: `snprintf` is 6.9% inclusive of `narrac emit selfhost` and `__vfprintf` 2.5% self, called through `Int.toString` from `effects.positionKey`, `effects.functionKey`, `emitstate.localName`, `emitstate.temp` and `cnames.internedName`. Profile proposal 1, saves about 5%. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] A sample printing `Int.toString` of 0, negative numbers, the minimum and maximum `int` and every unsigned width matches the old output.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

S: one file, a measured hotspot.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
