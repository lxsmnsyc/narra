---
id: missing-return
title: "Missing return is a compile error"
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
    why: builds on the compound assignment, C-style for and std modules batch-2 added
---

# Missing return is a compile error

## Goal

D92: a function with a result must return on every path. A path ends at `return`, a call to `panic`, an `if`/`else` whose branches all end, an exhaustive `match` whose arms all end, or `while (true)` / `for (;;)` with no `break` out of it. Today `function f(x: int): int { if (x > 0) { return 1 } }` checks clean and `f(0)` panics. Keep the emitted check as an internal compiler error. Error tests for each path form; samples for the forms that end.

## Notes

Outline, from D92 (maintainer, 2026-11-09). `selfhost/src` must follow the new rule in the same node: fix every site it rejects. Update SPEC.md (sections 2, 4 and 6 as needed) and drop "Not built yet" from D92 for this part.
