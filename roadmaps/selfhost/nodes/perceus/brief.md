---
id: perceus
title: Remove unneeded reference count updates
status: todo
priority: should
type: feature
outline: true
milestone: m6-performance
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
  - on: bench-harness
    kind: hard
    why: needs baseline numbers to show the gain
  - on: box-recursive
    kind: hard
    why: changes the heap layout this pass works on
---

# Remove unneeded reference count updates

## Goal

Perceus-style removal of unneeded increments and decrements, in-place reuse, and `--report-copies`, measured with `bench/`.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
