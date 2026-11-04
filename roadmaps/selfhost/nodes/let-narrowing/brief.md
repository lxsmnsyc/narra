---
id: let-narrowing
title: "Narrowing a let after an is test"
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

# Narrowing a let after an is test

## Goal

SPEC lines 407 and 569 (F9). After `if (x is None) { return 0 }`, `x.value` fails on a `let` variable with `this Option<int> may hold any of its variants`; it works on `const` and parameters. Narrow a `let` until its next assignment. Checker only.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
