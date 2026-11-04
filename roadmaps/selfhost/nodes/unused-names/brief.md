---
id: unused-names
title: "Unused locals and parameters are errors"
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

# Unused locals and parameters are errors

## Goal

SPEC line 57 (F5). Unused locals and parameters compile clean; the maintainer chose the spec over lint (2026-11-04), so they become checker errors with a fix (rename to `_name` or remove). Every sample in `compiler/testdata/` and every file in `selfhost/src` must be cleaned first, and the bootstrap fixed point must still hold. Checker only.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
