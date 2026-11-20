---
id: std-doc-comments
title: "Descriptions for every public std export"
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
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-doc-comments/**
shared:
reserves: []
provides: []
depends_on:
  - on: std-reference
    kind: hard
    why: the reference shows them
---
# Descriptions for every public std export

## Goal

One-line descriptions for every public std export: doc comments in `std/src/`, a description field in the builtin registry; the reference and the LSP hover show them. Proposal change 12.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
