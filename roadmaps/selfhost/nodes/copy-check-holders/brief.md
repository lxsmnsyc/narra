---
id: copy-check-holders
title: "The debug copy check counts copy credits per holder"
status: todo
priority: should
type: fix
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
spent:
owns:
  - roadmaps/selfhost/nodes/copy-check-holders/**
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimecore.nra
shared:
  - selfhost/tests/src/copies.nra
  - testdata/**
reserves: []
provides: []
depends_on: []
---

# The debug copy check counts copy credits per holder

## Goal

Found by shared-write-error (2026-11-11). The debug copy check (`NARRA_CHECK_COPIES=1`, `selfhost/runtime/runtime.h` near line 175) gives a payload box credits for the copies that `copy(...)` explains. It counts them on the box, not on its holders. When the original holder releases first, one copy's credits end early, and a later legitimate copy can be reported (exit 103). Track credits per holder so that only copies that no `copy(...)` explains are reported.

`selfhost/src/runtimecore.nra` is generated from `runtime.h`: rewrite it with `sh scripts/test.sh --suite generated --update`. Do not regenerate the seed; the orchestrator does that at the end of the batch.

## Acceptance

- Proof on `feat/batch-10`: a sample whose original holder releases before a legitimate copy, reported with exit 103 under `NARRA_CHECK_COPIES=1`, recorded in `log.md`.
- The same sample exits 0 on the branch, and a sample with an unexplained copy still exits 103. Both run in the `copies` suite.
- The release build is unchanged: the copy check code compiles out under `-DNARRA_RELEASE`.
- `sh scripts/test.sh` passes.
