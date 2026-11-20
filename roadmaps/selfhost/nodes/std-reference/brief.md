---
id: std-reference
title: "Generated std reference"
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
  - roadmaps/selfhost/nodes/std-reference/**
shared:
reserves: []
provides: []
depends_on:
  - on: std-interface-query
    kind: hard
    why: the pages come from its output
  - on: site-reference-section
    kind: hard
    why: the pages live in its section
---
# Generated std reference

## Goal

`/reference/std`: one page per public std module from `narrac query --json interface`, committed under `docs/reference/std/`, with a test that fails when a page is stale (the Vercel build may lack `cc`). Proposal change 11.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
