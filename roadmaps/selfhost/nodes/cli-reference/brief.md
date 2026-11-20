---
id: cli-reference
title: "Commands reference and complete usage text"
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
  - roadmaps/selfhost/nodes/cli-reference/**
shared:
reserves: []
provides: []
depends_on:
  - on: site-reference-section
    kind: hard
    why: the page lives in its section
---
# Commands reference and complete usage text

## Goal

One complete usage text in `narrac` covering `fmt`, `query`, `lsp`, `apply-fix`, `emit`; bare `narrac` prints it. `/reference/commands` by hand, with a test that every command and flag in `narrac --help` appears. Chapter 12's table becomes a link. Proposal change 13.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
