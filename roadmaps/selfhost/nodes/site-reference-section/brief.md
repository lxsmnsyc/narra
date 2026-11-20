---
id: site-reference-section
title: "Reference section on the site"
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
  - roadmaps/selfhost/nodes/site-reference-section/**
shared:
reserves: []
provides: []
depends_on:
  - on: first-release
    kind: hard
    why: after the release (docs-structure-review proposal, approved 2026-11-20)
---
# Reference section on the site

## Goal

`/reference` index, header link (Install, Guide, Reference), routes in `website/src/pages.ts`, search sources in `website/plugins/search.ts`, `docs/reference` in the `ignoreCommand` of `website/vercel.json`. Proposal change 9.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
