---
id: type-of
title: "typeof in type positions"
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
size: S
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
    why: builds on the batch-2 checker
---

# typeof in type positions

## Goal

D96: `typeof name` in a type annotation gives the declared or inferred type of a value in scope, with effects and `from` for functions. Names only; self-reference is an error. Parser, checker, `fmt`, `astdump`, `narra query` signatures, samples and error tests.

## Notes

Outline. Refine from the decision record named above.
