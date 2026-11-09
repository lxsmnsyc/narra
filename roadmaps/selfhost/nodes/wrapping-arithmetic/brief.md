---
id: wrapping-arithmetic
title: "Integer overflow wraps in every build"
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

# Integer overflow wraps in every build

## Goal

D92: `+`, `-`, `*` and unary `-` wrap in every build, so the debug overflow panics go. `MIN / -1` gives `MIN`, `MIN % -1` gives `0`. Remove `+%`, `-%`, `*%` and `+%=`, `-%=`, `*%=` (a `narrac fix` or a clear error pointing to the plain operator). Add checked functions that return `Option` (`Int.addChecked`, `subChecked`, `mulChecked`, and the sized-type versions).

## Notes

Outline, from D92 (maintainer, 2026-11-09). `selfhost/src` must follow the new rule in the same node: fix every site it rejects. Update SPEC.md (sections 2, 4 and 6 as needed) and drop "Not built yet" from D92 for this part.
