---
id: sized-numbers
title: "Sized integers, f32 and bytes"
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

# Sized integers, f32 and bytes

## Goal

SPEC lines 131 and 231-238 (F7). `i8` to `u64`, `i32`, `f32` and `bytes` are unknown types, and "a literal must fit its type" holds only for `int`. Types, literal range checks, conversions (`I32.from` and friends, line 534; see the spec-fixes outcome). Split when refined: integer types, then `f32`, then `bytes`.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- SPEC.md line 534 names `Int.toI32(x)` as a placeholder (spec-fixes, PR #59). Choose the real conversion names and fix that line.
- Outline. Break into a full brief when `plan.py refine` lists it.
