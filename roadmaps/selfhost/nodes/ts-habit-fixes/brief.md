---
id: ts-habit-fixes
title: "Diagnostics with fixes for every TypeScript habit"
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

# Diagnostics with fixes for every TypeScript habit

## Goal

SPEC lines 165-183, 1039, 191 (F16, F17). `class`, `interface`, `new`, `try`, `throw` give generic parse errors (E0019, E0013); `for in` (E0031), `number` (E0113), `'abc'` (E0002) and `.length` (E0208) have messages but no fix; `Arrays.push` without the import gives E0308 with no add-import fix (only `Array.push` gets E0908). Each row of the spec's table gets a diagnostic with a fix, with `testdata/diagnostics/` cases.

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Outline. Break into a full brief when `plan.py refine` lists it.
