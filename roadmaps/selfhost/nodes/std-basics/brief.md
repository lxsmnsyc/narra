---
id: std-basics
title: Math, time, random and hashing
status: todo
priority: should
type: feature
outline: true
milestone: m10-std
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests (std-source gate, 2026-11-05)
  - on: std-source
    kind: hard
    why: the spike decides whether this module is Narra source or compiler builtins
---

# Math, time, random and hashing

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

`std/math` (abs, min, max, clamp, pow, sqrt and the float functions in C's math.h), `std/time` (monotonic clock, wall clock, durations, sleep), `std/random` (a seeded generator and a system-seeded one) and `std/hash` (the existing `Hash.of`, plus a stable hash for files). Also `Ranges.stepped`, which `SPEC.md` documents but no module registers (inbox, 2026-10-30).

## Notes

Outline. May split by module when refined; each module is S.
- Deferred from std-foundation (2026-11-05): when a std source module first imports another one, sort std source modules with `orderModules` in single-file packages (`compiler/internal/narra/modules.go` and `selfhost/src/modules.nra`); today they skip it.
