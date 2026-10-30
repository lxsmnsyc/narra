---
id: selfhost-emit
title: Self-hosted C emitter
status: todo
priority: must
type: feature
outline: true
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: emits from the checked tree
---

# Self-hosted C emitter

## Goal

C emitter and driver in Narra, producing the same C as the Go compiler.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
