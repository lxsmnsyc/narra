---
id: typed-calls-and-effects
title: "Explicit type arguments in calls, and effects of function values"
status: todo
priority: could
type: feature
story:
early: false
outline: true
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Explicit type arguments in calls, and effects of function values

## Goal

SPEC lines 702, 791 and 805 (F20, F21, both in the README's "Not supported yet"). Calls through function values count as pure (an exported function calling `h.f()` does `io` without `uses io`), and `id<int>(1)` parses as comparisons. Parser (the balanced `<...>` rule of literal type arguments) and checker (effects on function types). May split in two.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
