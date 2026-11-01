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
  - on: std-source
    kind: hard
    why: the spike decides whether this module is Narra source or compiler builtins
---

# Math, time, random and hashing

## Goal

`std/math` (abs, min, max, clamp, pow, sqrt and the float functions in C's math.h), `std/time` (monotonic clock, wall clock, durations, sleep), `std/random` (a seeded generator and a system-seeded one) and `std/hash` (the existing `Hash.of`, plus a stable hash for files). Also `Ranges.stepped`, which `SPEC.md` documents but no module registers (inbox, 2026-10-30).

## Notes

Outline. May split by module when refined; each module is S.
