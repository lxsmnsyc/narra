---
id: batch-3
title: "Batch 3: the compiler batch"
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
  - roadmaps/selfhost/nodes/batch-3/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 3: the compiler batch

## Goal

Third one-shot batch, scope chosen by the maintainer on 2026-11-09: the nodes that make the self-hosted compiler's own code clearer, safer or faster. Children land on `feat/batch-3` as direct commits after targeted tests; at the end one full check, one Linux check, one fresh review and one pull request.

Children: or-patterns, missing-return, wrapping-arithmetic, type-aliases, decls-quadratic, template-positions, lsp-navigation-parse, unused-lines, then adopt-new-syntax last.

## Seed

adopt-new-syntax uses or-patterns and compound assignment in `selfhost/src`. The orchestrator regenerates the seed after the other children land and before adopt-new-syntax starts, then once more at the end.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] The inbox entries the fixes answer are removed from `inbox.md`.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [ ] The seed is regenerated at the end and `sh selfhost/bootstrap.sh` reports it current.
