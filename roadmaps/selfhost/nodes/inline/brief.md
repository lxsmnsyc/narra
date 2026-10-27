---
id: inline
title: @inline for higher-order functions
status: todo
priority: could
type: feature
outline: true
milestone: m6-performance
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
    why: inlining a higher-order function needs closures
---

# @inline for higher-order functions

## Goal

`@inline` copies a function's body and its closure argument into the caller.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
