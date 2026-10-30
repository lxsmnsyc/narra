---
id: perceus
title: Remove unneeded reference count updates
status: todo
priority: later
type: feature
outline: true
milestone: m6-performance
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: bench-harness
    kind: hard
    why: needs baseline numbers to show the gain
  - on: box-recursive
    kind: hard
    why: changes the heap layout this pass works on
---

# Remove unneeded reference count updates

## Goal

Perceus-style removal of unneeded increments and decrements, in-place reuse, and `--report-copies`, measured with `bench/`.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.

- box-recursive (PR #10, 2026-10-30) put every fielded enum variant on the heap, as SPEC section 6 says. `bench/run.sh -n 9 interp` went from about 0.20 s on main to about 0.41 to 0.49 s. The hot loop retains and releases a heap `Op` 150M times without ever dropping it. Borrowing the loop variable (or iterating `Array.values` by borrow) should win most of it back. Bring `interp` back near its pre-box-recursive time as one acceptance criterion. The user accepted the regression until this node lands.
