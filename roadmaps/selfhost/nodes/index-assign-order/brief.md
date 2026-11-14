---
id: index-assign-order
title: "Indexed assignment evaluates its right side first"
status: done
priority: must
type: fix
story: batch-15
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent: 203000
owns:
  - roadmaps/selfhost/nodes/index-assign-order/**
  - testdata/index_assign_order*
shared:
  - selfhost/src/*.nra
reserves: []
provides: []
depends_on: []
---
# Indexed assignment evaluates its right side first

## Goal

`grid[idx(0)][0] = loud(9)` calls `loud` before `idx`; with `= await step()` the target is evaluated after the await. TypeScript evaluates the target (object and index expressions, left to right) first, then the right side, then stores. Match that, for plain, compound (`+=`) and nested indexed and field targets, and in async functions across an `await`.

## Context

- The emitter for assignment statements (`grep -rn 'IndexAssign\|emitAssign' selfhost/src | head`) and the async lowering, which must keep evaluated target parts in the frame across a suspension.
- Unique-array and copy-on-write paths (`loop-unique-hoist`, the array write cold paths) must stay correct: the store still makes the array unique at the store, after the right side runs.
- Watch cost: record `emit selfhost` instructions; simple targets (local array, constant index) should emit as before.

## Acceptance

- [ ] `testdata/index_assign_order.nra` fails on the base and passes: printed call order for plain, compound, nested, map index, field-then-index, and an async case with `await` on the right.
- [ ] Clean under the copy check and AddressSanitizer. `sh scripts/test.sh` passes.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
