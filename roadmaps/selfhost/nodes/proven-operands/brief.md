---
id: proven-operands
title: "Divisors and shift counts proven by narrowing"
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
    why: builds on the compound assignment, C-style for and std modules batch-2 added
---

# Divisors and shift counts proven by narrowing

## Goal

D92: integer `/` and `%` compile only when the divisor is a non-zero constant or narrowed by a check (`!= 0`, `> 0`, `< 0`). Shift counts must be a constant in `[0, width)` or narrowed (`n >= 0 && n < 64`; only the upper bound for unsigned). Otherwise a compile error with a hint. The runtime division-by-zero and shift-count panics go. Reuse the narrowing of `let` and `is` (let-narrowing); decide how a narrowing ends when the name is reassigned, and whether a field path (`size.count`) can be narrowed.

## Notes

Outline, from D92 (maintainer, 2026-11-09). `selfhost/src` must follow the new rule in the same node: fix every site it rejects. Update SPEC.md (sections 2, 4 and 6 as needed) and drop "Not built yet" from D92 for this part.
