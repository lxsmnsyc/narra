---
id: copy-builtin
title: "copy(x) and the debug copy check"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 runtime and Cell
---

# copy(x) and the debug copy check

## Goal

`copy(x)` at any depth as a retain whose result the checker treats as unshared. Debug builds keep a credit set (P12, P18, O18): `copy` grants a credit, copy-on-write uses one, debug-only hooks at every reference-count decrement drop entries; any copy-on-write without a credit is reported at exit. Remove `@unique` from SPEC and its skipped sample.

## Notes

Outline, from D94 and `roadmaps/selfhost/debate/value-semantics/debate.md` (step 1 of 6). Refine with the debate's claims named here.
