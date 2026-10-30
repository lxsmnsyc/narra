---
id: regions
title: Regions
status: todo
priority: later
type: feature
outline: true
milestone: m4-systems
target:
from:
size: M
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
  - on: closures
    kind: hard
    why: Region.run takes a closure
---

# Regions

## Goal

`Region.run` with bump allocation, copying values out, and no reference count updates inside.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
