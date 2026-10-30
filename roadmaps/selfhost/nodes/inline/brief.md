---
id: inline
title: @inline for higher-order functions
status: todo
priority: later
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
- 2026-10-30: the user wants better closure management once std/option and std/result add more higher-order helpers; std-option-result records the cost per call to measure against.
- From closures: a non-escaping closure gets a stack environment, which is a block-local C variable. Inlining or any tail-call rewrite must not move a pointer to it out of its block.
