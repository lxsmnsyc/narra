---
id: batch-10
title: "Batch 10: hidden copies, batch-9 leftovers and the destructuring design"
status: done
priority: should
type: story
story: 
early: false
outline: false
milestone: m9-adoption
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-12
spent: 1515000
owns:
  - roadmaps/selfhost/nodes/batch-10/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-9
    kind: hard
    why: the inbox items below are findings of batch-9 and its review
---

# Batch 10: hidden copies, batch-9 leftovers and the destructuring design

## Goal

Tenth one-shot batch, scope set by the maintainer on 2026-11-12: the hidden copies that keep the debug copy check out of CI, the leftovers from the batch-9 reviews, and the destructuring design spike. Children land on `feat/batch-10` as direct commits after a worker check and a fresh review. At the end: two seed updates, the full check and the Linux check, one fresh review of the whole branch, a bench row, and one pull request.

Children:

- checker-emitter-copies: a debug-built compiler compiles `selfhost` with `NARRA_CHECK_COPIES=1` and exits 0. Close the E0978 gaps behind the copies (a write after another name's last use, and array elements that share one value), then fix the patterns in `selfhost/src`.
- copy-check-holders: the debug copy check counts credits per holder, not per payload box, so a legitimate copy is not reported after the original holder releases first.
- fmt-std-sources: the `sources` suite checks `std/src`, the unformatted std files get formatted, and `endsWithBlock` looks at every operand of an assignment chain.
- seed-switch-cleanup: `narrac run` exits through `exitWithoutRelease` when `startProgram` fails, and `buildtarget.nra` matches the host target by its `Os` and `Arch` values, not by text.
- destructuring-design: decision D103 on destructuring in parameters, array patterns and assignments, and refined briefs for its three children (gate: the maintainer approves D103).

Dropped from the inbox while planning, as stale:

- `TestSelfhostBuiltins` signature check: a Go test, gone with the Go compiler (story c-seed).
- The skipped `@unique` SPEC sample (audit finding F15): D94 removed `@unique`, and SPEC has no such sample now.

## Order and shared files

- One wave, five workers, on an 8 GB machine: start checker-emitter-copies, copy-check-holders, fmt-std-sources and destructuring-design, then seed-switch-cleanup when a slot frees.
- checker-emitter-copies edits code across `selfhost/src`. The S nodes that touch `driver.nra`, `buildtarget.nra` and `fmtexprs.nra` land first when they are ready, and checker-emitter-copies rebases onto them. It reruns `narrac fmt` after the rebase.
- copy-check-holders changes what the copy check reports. If it lands first, checker-emitter-copies measures against it. Either way the orchestrator reruns the debug-built `emit selfhost` on the batch after both land.
- Each worker uses a scratch folder named after its node under /private/tmp/claude-501/. Check a rebase's exit code before testing.

## Seed

copy-check-holders edits `selfhost/runtime/runtime.h`, and fmt-std-sources edits std text, so the end of the batch is two seed updates (GOAL.md constraint). No child uses a feature the current seed lacks, so there is no seed update in the middle of the batch.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [x] A debug-built compiler (seed built without `-DNARRA_RELEASE`) runs `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0 on the story branch.
- [x] A fresh review of the whole story branch, before the pull request.
- [x] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise by more than the noise.
- [x] The inbox entries this batch closes, and the two stale ones above, are removed from `inbox.md`.
