---
id: const-lazy
title: "lazy values and @const functions"
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

# lazy values and @const functions

## Goal

SPEC lines 910-915 (F12, in the README's "Not supported yet"). `const table = lazy build()` and `@const function` fail to parse. Parser, checker and emitter (a once guard for `lazy`; compile-time evaluation for `@const`).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
