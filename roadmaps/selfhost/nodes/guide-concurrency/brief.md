---
id: guide-concurrency
title: "Guide chapter on async, threads and channels"
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
  - roadmaps/selfhost/nodes/guide-concurrency/**
shared:
reserves: []
provides: []
depends_on:
  - on: first-release
    kind: hard
    why: after the release (docs-structure-review proposal, approved 2026-11-20)
---
# Guide chapter on async, threads and channels

## Goal

A guide chapter on async, threads and channels, which the guide index lists as built but does not teach. Proposal change 17.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
