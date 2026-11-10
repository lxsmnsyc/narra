---
id: regions
title: Regions
status: todo
priority: should
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
  - on: regions-design
    kind: hard
    why: D100 decides the design and splits this node into children
---

# Regions

## Goal

`Region.run` with bump allocation, copying values out, and no reference count updates inside.

## Notes

Outline until regions-design writes D100; refine the children from it.
