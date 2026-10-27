---
id: selfhost-bootstrap
title: Bootstrap: the compiler compiles itself
status: todo
priority: must
type: feature
outline: true
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: true
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-emit
    kind: hard
    why: needs a complete compiler in Narra
---

# Bootstrap: the compiler compiles itself

## Goal

Stage 1 (built by Go) builds stage 2 (built by stage 1). Both produce byte-identical C on all test programs. Document the bootstrap steps.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
