---
id: dynamic-traits
title: "Dynamic<Trait> values"
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
size: L
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

# Dynamic<Trait> values

## Goal

SPEC lines 501-503 (F14, in the README's "Not supported yet"). `Dynamic.of(Circle { r: 1 })` fails to parse. Parser, checker and emitter (a boxed value with a table of trait functions).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.

- 2026-11-09: design settled in D95 (maintainer). Refine the brief from it.
