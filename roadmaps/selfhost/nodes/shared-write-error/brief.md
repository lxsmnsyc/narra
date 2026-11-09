---
id: shared-write-error
title: "Shared writes become errors"
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
  - on: shared-write-sweep
    kind: hard
    why: the compiler must be clean first
---

# Shared writes become errors

## Goal

Turn the warning into an error, turn on the debug copy check in the test runs, and add an error sample for each D2 form (P20 step 5).

## Notes

Outline, from D94 and `roadmaps/selfhost/debate/value-semantics/debate.md` (step 5 of 6). Refine with the debate's claims named here.
