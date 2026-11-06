---
id: typed-calls-and-effects
title: "Effects of function values"
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

# Effects of function values

## Goal

SPEC line 702 (F20, in the README's "Not supported yet"). Calls through function values count as pure (an exported function calling `h.f()` does `io` without `uses io`). Checker work: effects on function types. Explicit type arguments in calls (F21) moved to node call-type-args (2026-11-07).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
