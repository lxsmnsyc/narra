---
id: simd
title: SIMD vector types
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
  - on: fixed-arrays
    kind: hard
    why: vectors build on fixed-size array layout
---

# SIMD vector types

## Goal

`std/simd` vector types compiled to C vector extensions.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
