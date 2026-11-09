---
id: shared-write-check
title: "Shared-write analysis, as a warning"
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
size: L
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
  - on: copy-builtin
    kind: hard
    why: the fix the warning offers is copy(x)
  - on: from-marker
    kind: hard
    why: call results share their from arguments
---

# Shared-write analysis, as a warning

## Goal

The analysis of P17 with O17: tracked types, conflicts with any live name sharing the root or a containing path, two-way chains, `mutable R` writes below R, closure writes at the pass site, check before the strong update. Ships as a warning with `narrac apply-fix` fixes, so the real site count is measured before code must change.

## Notes

Outline, from D94 and `roadmaps/selfhost/debate/value-semantics/debate.md` (step 3 of 6). Refine with the debate's claims named here.
