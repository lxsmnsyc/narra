---
id: for-of-sources
title: "for-of over any Range, Array or Str.chars value"
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

# for-of over any Range, Array or Str.chars value

## Goal

SPEC lines 110-111 (F6). `for ... of` accepts only direct calls to `Ranges.upTo`, `Ranges.between`, `Arrays.values`, `Maps.keys` and `Maps.entries`. Accept a `Range` held in a variable, an array value directly, and `Str.chars`. `Arrays.entries` is added by for-destructure, not here.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Found in the for-destructure review: `Arrays.values` and `Arrays.entries` used outside a for-of panic the Go emitter (`emit: unknown builtin`) instead of giving a compile error. Fix that here.
- Outline. Break into a full brief when `plan.py refine` lists it.
