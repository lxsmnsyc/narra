---
id: spec-samples-test
title: "Every SPEC.md sample is checked by a test"
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
  - on: spec-fixes
    kind: hard
    why: the spec text must be right before samples become tests
---

# Every SPEC.md sample is checked by a test

## Goal

Audit learning: F4 and F8 were bugs in features marked done. A test extracts the fenced `ts` blocks from SPEC.md and runs `narra check` on each (wrapping fragments in a function where needed), with a skip list of samples that need planned features. Removing a skip is part of each fix node from then on, like the emit skip files were.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
