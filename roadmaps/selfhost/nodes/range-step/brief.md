---
id: range-step
title: "Ranges.stepped(start, end, step)"
status: todo
priority: must
type: feature
story: batch-15
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/range-step/**
  - testdata/range_step*
  - testdata/errors/range_step*
shared:
  - selfhost/src/*.nra
  - std/src/iterator.nra
  - docs/guide/*.md
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1270-E1279]
provides: []
depends_on: []
---
# Ranges.stepped(start, end, step)

## Goal

D119: `Ranges.stepped(start, end, step)` for counting down and strides. End is exclusive as in the plain range. A negative step counts down (`stepped(10, 0, -2)` gives 10, 8, 6, 4, 2). A constant step of 0 is a compile error (code from E1270-E1279); a run-time step of 0 gives an empty range. No overflow at the end of the int range.

## Context

- Find the range module with `grep -rn 'Ranges' selfhost/src | head`. for-of over a range is lowered to a C loop; `stepped` should lower the same way, not through a generator.
- Works as a value passed to std/iterator operators, like the plain range.
- `@const` bodies may call it if plain ranges are allowed there.

## Acceptance

- [ ] `testdata/range_step.nra` with output: up, down, stride past the end, empty, run-time zero, near `Int.max` and `Int.min`, through `Iterators.map`.
- [ ] Error test for a constant zero step. SPEC, guide and std docs list it.
- [ ] `sh scripts/test.sh` passes.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
