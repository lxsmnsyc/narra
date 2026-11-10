---
id: loop-unique-hoist
title: "Make a local array unique once before a loop that writes it by index"
status: active
priority: should
type: feature
story: batch-5
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/loop-unique-hoist/**
shared:
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/exprs.nra
  - testdata/**
reserves: []
provides: []
depends_on:
  - on: array-write-cold-paths
    kind: hard
    why: builds on the split `_unique` and `narra_bounds`
---

# Make a local array unique once before a loop that writes it by index

## Goal

When a `while`, C-style or range `for` body writes `a[i] = ...` or `a[i].field = ...` to a local array, and the body uses `a` only as `a[expr]` reads, index writes and `Arrays.length(a)` (no copy into another binding, no other argument use, no closure capturing `a`), and the loop is not inside a `function*` body, emit `a = A_unique(a);` once before the loop and emit each write as `narra_bounds(i, a->len, where); &a->items[i]` with no `rc` test.

## Context

- Evidence: particles keeps the `rc == 1` test and the `_copy` call in its loop, so each element reloads the array header, about 26 instructions per particle. Rewriting the emitted C this way took particles from 0.107 s to 0.068 s (2.61G to 1.51G instructions, ahead of Go 0.074 s). Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-programs/outcome.md`, proposal 2, and `experiments/variant3.py`.
- `emitReferenceFor` in `emitstmts.nra` already does this for `for (mutable x of items)`; `checkLoopBorrows` in `exprs.nra` has a related walk to reuse.
- Generators are excluded: a generator frame holds the array across `yield`, and copying the generator (`let h = g`) retains it, so hoisted writes would change storage both copies share. The compiler refuses `for (mutable ...)` in a `function*` for the same reason (E0957).

## Acceptance

- [ ] particles and sieve times and instructions before and after, with the particles gain above the spread.
- [ ] A sample with a `function*` that writes `a[i] = 1` in a loop and yields the sum after each write, resumed once, copied (`let h = g`), then both copies resumed, prints `1 2 3 2 3`.
- [ ] A sample where the loop body copies the array into another binding, or passes it to a call, keeps the per-write check and prints the same output as on the base.
- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. Report the change; no gain is required here.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
