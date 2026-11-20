---
id: manifest-reference
title: "narra.toml reference"
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
  - roadmaps/selfhost/nodes/manifest-reference/**
shared:
reserves: []
provides: []
depends_on:
  - on: site-reference-section
    kind: hard
    why: the page lives in its section
---
# narra.toml reference

## Goal

`/reference/narra-toml`: every table and key, type, default, and the error code for a wrong value, gathered from chapters 1, 8, 12 and `SPEC.md` sections 2 and 15. Proposal change 14.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
