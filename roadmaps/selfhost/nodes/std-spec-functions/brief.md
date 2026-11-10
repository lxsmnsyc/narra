---
id: std-spec-functions
title: "Std functions and constants the spec uses"
status: todo
priority: should
type: feature
story: batch-6
early: false
outline: false
milestone: m10-std
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
owns:
  - testdata/std_spec_*
  - roadmaps/selfhost/nodes/std-spec-functions/**
shared:
  - selfhost/src/std*.nra
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/spec-skip.txt
  - selfhost/tests/src/**
  - testdata/**
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests
---

# Std functions and constants the spec uses

## Goal

Finding F18 of the spec-gap audit (`nodes/spec-gap-audit/outcome.md`, findings table), except `Arrays.entries` (built). Check each name against the current tree first; some exist now and some conflict with later decisions:

- Missing on 2026-11-11: `Float.PI` (std modules have no constants yet: add module constants to the checker, or write the constant in a source module as a top-level `const` if top-level-const allows it), `Int.divmod`, `Int.floorDiv`, `Int.compare`, `Str.find`, `Arrays.slice`, `Arrays.withCapacity`.
- Check before adding: `Arrays.map` (the 3. Modules sample in `selfhost/tests/spec-skip.txt` says std/array has none), `Int.truncate` (exists in std/float and std/f32 as a conversion; the spec may mean something else).
- Do not add `Str.copy` or `Arrays.copy`: D94 makes `copy(x)` the one copy tool and says `Arrays.copy` stops being one. Change the spec samples that use them to `copy(...)`.
- Results follow D92: no function added here panics. `Int.divmod` and `Int.floorDiv` take a divisor; decide whether they return `Option` or need a proven divisor like `/` (proven-operands), and record the choice.

Remove the spec-skip lines these functions unblock (`Float.PI`, `std/array has no map`, `Int.compare`).

## Acceptance

- [ ] Each added function or constant has a program sample.
- [ ] The three spec-skip lines are gone and the spec suite passes.
- [ ] SPEC.md has no `Str.copy` or `Arrays.copy`.
- [ ] `sh scripts/test.sh` passes.
