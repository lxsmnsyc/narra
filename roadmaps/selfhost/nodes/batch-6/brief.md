---
id: batch-6
title: "Batch 6: shared writes and the D92 proofs"
status: done
priority: should
type: story
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/batch-6/**
shared: []
reserves: []
provides: []
depends_on:
  - on: batch-5
    kind: hard
    why: builds on copy-builtin, from-marker and the seed regenerated at the end of batch-5
---

# Batch 6: shared writes and the D92 proofs

## Goal

Sixth one-shot batch, scope set by the maintainer on 2026-11-11: the rest of D94 (debate `roadmaps/selfhost/debate/value-semantics/debate.md`, steps 3 to 6), the unbuilt parts of D92, D96 (`typeof`), the spec's missing std functions, and the inbox follow-ups from batch-5. Children land on `feat/batch-6` as direct commits after targeted tests. At the end the batch gets one full check, one Linux check, one seed update, one fresh review of the whole branch and one pull request.

Children:

- D94 in order (P20): shared-write-check, shared-write-sweep, shared-write-error, value-semantics-docs. Only one of them runs at a time.
- D92: proven-operands and std-no-panic. They run before shared-write-sweep, because all three rewrite call sites across `selfhost/src` (the debate orders D94 after the D92 nodes).
- D96: type-of.
- F18: std-spec-functions.
- Inbox: from-marker-followups, unreachable-arms, std-compound-sweep (after std-no-panic, both edit `std/src`).

## Checkpoint after shared-write-check

shared-write-check ships the rule as a warning and reports how many sites `selfhost/src`, `std/src` and the tests have. The debate estimated 120 to 220 edits and 30 to 50 `from` marks; from-marker found 185 shared returns. Before shared-write-sweep starts, the orchestrator posts the counts to the maintainer and splits the sweep if it is larger than L (for example by directory: checker, emitter, tools, std and tests).

## Seed

batch-5 ended with a seed update, so the seed reads `copy(...)` and `from`. shared-write-sweep can mark the compiler without a seed update first. proven-operands changes what the compiler accepts in `selfhost/src` and fixes the sites it rejects in the same commit; the fixes add checks, which the current seed reads. std-no-panic changes the result types of builtins the compiler may call. The seed's tables have the old types, so compiler call sites stay on the old form until a seed update; if there are any, the end of the batch is seed update, switch those sites, seed update. Otherwise regenerate the seed once at the end.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated at the end and `sh selfhost/bootstrap.sh` reports it current.
- [x] The textops helpers (`substring`, `splitString`, `repeatString`, `filledArray` in `selfhost/src/textops.nra`) call the Option builtins directly, and the seed is regenerated again. The seed shared-write-sweep regenerated mid-batch (e61449e6) already reads the Option builtins, so the switch needs only the end-of-batch seed update after it. Until then the compiler costs about 32% more instructions (std-no-panic review, 2026-11-11). The end-of-batch bench row must show it gone.
- [x] A fresh review of the whole story branch, before the pull request (wave-18 lesson).
- [x] `bench/BASELINE.md` has an end-of-batch row, recorded on an idle machine (the sweep adds `copy(...)` calls to the compiler).
