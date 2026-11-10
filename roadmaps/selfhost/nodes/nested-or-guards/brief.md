---
id: nested-or-guards
title: "Guards try every alternative of a nested or-pattern"
status: todo
priority: should
type: fix
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - testdata/or_patterns_nested_guard*
  - roadmaps/selfhost/nodes/nested-or-guards/**
shared:
  - selfhost/src/emitmatch.nra
  - roadmaps/selfhost/inbox.md
reserves: []
provides: []
depends_on: []
---

# Guards try every alternative of a nested or-pattern

## Goal

A guard tries each matching alternative only for an or-pattern at the top of a `match` arm (batch-3 review). An or-pattern nested inside another pattern, such as `Some(A(x) | B(x)) if x > 1`, still binds from the first matching alternative, so the arm can be skipped when a later alternative would pass the guard. Make the nested case try every alternative, as the top-level case does.

## Context

- Top-level handling: `emitAlternativesCondition` and the first-alternative binding in `selfhost/src/emitmatch.nra:156-211`. Read the or-patterns outcome and the batch-3 review fix (commit 42258be) for how the top level was fixed.
- Proof first: a sample where the first matching nested alternative fails the guard and a later one passes; it takes the wrong arm on the base.

## Acceptance

- [ ] `testdata/or_patterns_nested_guard.nra` with nested alternatives at depth 1 and 2, with and without bindings.
- [ ] `sh scripts/test.sh` passes; inbox entry removed.

## Sizing

S: extends an existing fix to nested patterns; may be M if nested alternatives need backtracking across sibling patterns. Say so in the outcome if it grows.
