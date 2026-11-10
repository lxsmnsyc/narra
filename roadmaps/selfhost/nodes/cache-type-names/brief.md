---
id: cache-type-names
title: "Cache C type names in the emitter"
status: todo
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
  - roadmaps/selfhost/nodes/cache-type-names/**
shared:
  - selfhost/src/cnames.nra
  - selfhost/src/emitstate.nra
reserves: []
provides: []
depends_on:
  - on: perceus
    kind: hard
    why: both edit emitstate.nra; the earlier node lands first
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# Cache C type names in the emitter

## Goal

`cnames.typeId` and `cnames.cType` rebuild the name, its `xxh32` hash and `hex8` text on every call. Cache them per type key in the emitter. `Emitter.typeNames` (`selfhost/src/emitstate.nra:53`) is declared but never read or written; use it or remove it.

## Context

- Evidence: `typeId` 6.1% and `cType` 4.9% inclusive, from `temp`, `refCountCode`, `functionSignature` and `declareLocalVariable`. Profile proposal 4, saves about 4%. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] The emitted C is byte-identical (the bootstrap fixed point and the `emit` suite's cnames stability check).
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

S: one file, a measured hotspot.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
