---
id: emit-string-helpers
title: "Faster cString and writeLine in the emitter"
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
spent:
owns:
  - roadmaps/selfhost/nodes/emit-string-helpers/**
shared:
  - selfhost/src/emitstate.nra
reserves: []
provides: []
depends_on:
  - on: perceus
    kind: hard
    why: both edit emitstate.nra; the earlier node lands first
  - on: cache-type-names
    kind: hard
    why: both edit emitstate.nra; the earlier node lands first
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# Faster cString and writeLine in the emitter

## Goal

`cString` (`selfhost/src/emitstate.nra:198`) appends bytes directly; `findByteChar` makes a one-character `Str.slice` per byte today. `writeLine` appends the indent and the text to `emitter.out` without building `padding + text + "\n"` first.

## Context

- Evidence: `cString` 3.4% inclusive (`findByteChar` 1.6%) and `writeLine` 2.6%. Profile proposal 9, saves about 3%. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] The emitted C is byte-identical (bootstrap fixed point).
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

S: one file, a measured hotspot.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
