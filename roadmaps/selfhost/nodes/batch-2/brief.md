---
id: batch-2
title: "Batch 2: inbox fixes, loop and assignment syntax, Cell, std basics, style sweep"
status: todo
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
added: 2026-11-09
spent:
owns:
  - roadmaps/selfhost/nodes/batch-2/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 2: inbox fixes, loop and assignment syntax, Cell, std basics, style sweep

## Goal

Second one-shot batch, scope chosen by the user on 2026-11-09. Children land on `feat/batch-2` as direct commits after targeted tests. At the end: one full check, one regenerated seed, one fresh review and one pull request into main.

Children:

- Fixes from the inbox: template-nested-strings, effect-variable-values, generator-function-type, misleading-hints.
- Features: c-style-for, compound-assign, cell-capture, std-basics.
- spec-samples-test, after the features, so its skip list starts small.
- style-sweep last, because it touches every hot file.

Left for later: release-binaries, narra-skills, selfhost-layout.

## Gates

As in batch-1, the orchestrator makes the gate decisions of gated children (cell-capture, effect-variable-values), writes each to DECISIONS.md, and lists them at the top of the pull request. The user may reject any of them before merging. The story is gated, so the pull request waits for the user.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] The inbox entries the fixes answer are removed from `inbox.md`.
- [x] `sh scripts/test.sh` passes on the story branch, and `sh scripts/linux-check.sh` passes.
- [x] The seed is regenerated once at the end and `sh selfhost/bootstrap.sh` reports it current.
- [x] The pull request lists every gate decision the orchestrator made.

All ten children landed. Fresh review of the whole branch found six problems (Cells.update use-after-free, duplicate C types when generator effects differ, `+=` after a block-bodied closure, `+=` skipping the mutable conflict check, missing `alloc` on calls through generator-returning values, lost E0003 inside brackets), all fixed with tests. The seed was regenerated at the end and again after the review fixes. Final checks at 623f819: scripts/test.sh PASS, linux-check PASS, bootstrap reports the seed current.
