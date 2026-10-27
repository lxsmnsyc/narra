---
id: fixed-arrays
title: Fixed-size arrays and const generics
status: todo
priority: could
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
  - on: split-compiler
    kind: hard
    why: adds a type to the files split-compiler creates
---

# Fixed-size arrays and const generics

## Goal

`FixedArray<T, N>` stored inline, with `const N: int` type parameters.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
