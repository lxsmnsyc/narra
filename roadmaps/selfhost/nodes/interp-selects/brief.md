---
id: interp-selects
title: "Spike: the interp slowdown from inline enum payloads"
status: todo
priority: should
type: spike
story:
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: S
spike: true
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/interp-selects/**
shared: []
reserves: []
provides: []
depends_on:
  - on: batch-4
    kind: hard
    why: builds on the batch-4 compiler and its measurements
---

# Spike: the interp slowdown from inline enum payloads

## Goal

option-inline (D99) made `interp` about 20% slower in user time (0.15 s to 0.18 s): with the `Op` payload in the value, clang compiles the hot `Add`/`Sub` arms into conditional selects that lengthen the dependency chain. A `switch` rewrite and a payload pointer did not help. Find out whether a different emitted form for matches on inline enums avoids the selects, and whether other match-heavy code pays the same cost.

## Context

- Read `git show feat/batch-4:roadmaps/selfhost/nodes/option-inline/outcome.md` and D99. Compare clang and gcc output for the hot function.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [ ] `outcome.md` either names an emitted form that brings `interp` back to 0.15 s without slowing the compiler, as a proposed brief, or says why not.

