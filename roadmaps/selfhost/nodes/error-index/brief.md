---
id: error-index
title: "Generated error index"
status: todo
priority: could
type: feature
story: 
early: false
outline: true
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/error-index/**
shared:
reserves: []
provides: []
depends_on:
  - on: site-reference-section
    kind: hard
    why: the page lives in its section
---
# Generated error index

## Goal

`narrac explain --list --json` and a generated, committed `/reference/errors` (479 codes, 96 with explanations), checked like the std reference. Proposal change 15.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
