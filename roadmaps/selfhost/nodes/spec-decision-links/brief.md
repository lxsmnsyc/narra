---
id: spec-decision-links
title: "D-numbers on /spec link to DECISIONS.md"
status: todo
priority: could
type: docs
story: 
early: false
outline: true
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/spec-decision-links/**
shared:
reserves: []
provides: []
depends_on:
  - on: first-release
    kind: hard
    why: after the release (docs-structure-review proposal, approved 2026-11-20)
---
# D-numbers on /spec link to DECISIONS.md

## Goal

The 99 `D` numbers on `/spec` render as links to `DECISIONS.md` on GitHub (`website/plugins/markdown.ts`). Proposal change 16.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
