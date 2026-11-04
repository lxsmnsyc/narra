---
id: contracts
title: "requires and ensures contracts"
status: todo
priority: later
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

# requires and ensures contracts

## Goal

SPEC lines 1007-1016 (F11). `requires x >= 0.0` / `ensures result >= 0.0` fail to parse. Parser, checker (`result` name, bool type) and emitter (checks in debug builds only).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
