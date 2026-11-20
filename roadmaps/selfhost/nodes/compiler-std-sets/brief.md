---
id: compiler-std-sets
title: "The compiler uses std/set"
status: done
priority: should
type: refactor
story: batch-23
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/compiler-std-sets/**
shared:
  - selfhost/src/**
  - testdata/**
reserves: []
provides: []
depends_on: []
---
# The compiler uses std/set

## Goal

Batch-22 added `std/set` (`Set<T>`: `create`, `add`, `remove`, `has`, `size`, `values`, `union`, `intersection`, `difference`). `selfhost/src` has 155 `Map<K, bool>` declarations in 42 files; most are sets written as maps. Replace each one used only for membership (value always `true`, read with `Maps.has` or `Maps.get(...) == Some(true)`) with `Set<K>`. Leave a map whose `false` value means something, and list those in the outcome.

## Rules

- Behavior does not change: every golden and emitted C output stays the same. Iteration order of `Set.values` must match what the map gave where the order is visible (diagnostic order, emitted declarations); check the emitted C of selfhost is byte-identical before and after (`narrac emit selfhost`).
- The compiler must not get slower or use more memory: `Set<T>` wraps `Map<T, bool>`, so expect no change; compare `bench/run.sh -n 5 compiler` (time and peak memory) before and after, and report both with the load average.
- Work file by file; commit once at the end. No function-level comments, no `<noun>Of` names, no D-numbers in code.
- Another worker (compiler-std-sorts) edits the same files for sorts and character checks; touch only set code, so the two changes merge cleanly.

## Acceptance

- [ ] Every membership-only `Map<K, bool>` in `selfhost/src` is a `Set<K>`; the kept maps are listed with the reason.
- [ ] `narrac emit selfhost` output is identical before and after.
- [ ] `sh scripts/test.sh` passes; `sh selfhost/bootstrap.sh` reaches the fixed point; the compiler row is not slower or larger beyond noise.
