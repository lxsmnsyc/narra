---
id: type-aliases
title: "type and export type aliases"
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

# type and export type aliases

## Goal

SPEC lines 209 and 219 (F13, in the README's "Not supported yet"). `type Tok = Token` fails to parse. Aliases resolve to the target type, and enum constructors work through the alias. Parser and checker; emitter unchanged.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
