---
id: std-spec-functions
title: "Std functions and constants the spec uses"
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
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests (std-source gate, 2026-11-05)
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
  - on: std-source
    kind: hard
    why: std-source decides whether std modules become Narra source; new std functions should follow that decision
---

# Std functions and constants the spec uses

## Goal

F18 except `Arrays.entries`. Missing: `Float.PI` (std modules have no constants yet), `Int.divmod`, `Arrays.map`, `Int.floorDiv`, `Int.compare`, `Str.find`, `Str.copy`, `Arrays.slice`, `Arrays.copy`, `Arrays.withCapacity`, `Int.truncate` (check the spec-fixes outcome: some names may move).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
