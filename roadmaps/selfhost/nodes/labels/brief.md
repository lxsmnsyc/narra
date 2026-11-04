---
id: labels
title: "Labeled break and continue"
status: todo
priority: should
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

# Labeled break and continue

## Goal

SPEC line 120 (F3). `outer: for (...)` and `outer: while (true)` fail to parse. Parser, checker (label scope, unknown label, duplicate label) and emitter (`goto` past the loop or to its continue point).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
