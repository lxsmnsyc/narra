---
id: proven-operands
title: "Divisors and shift counts proven by narrowing"
status: todo
priority: should
type: feature
story: batch-6
early: false
outline: false
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
owns:
  - testdata/errors/divisor_*
  - testdata/errors/shift_*
  - testdata/proven_*
  - roadmaps/selfhost/nodes/proven-operands/**
shared:
  - selfhost/src/**
  - selfhost/runtime/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0979, diag E0980]
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the compound assignment, C-style for and std modules batch-2 added
---

# Divisors and shift counts proven by narrowing

## Goal

Build the unbuilt part of D92 (DECISIONS.md):

- Integer `/` and `%` (and `/=`, `%=`) compile only when the divisor is a non-zero constant or a name narrowed by a check (`!= 0`, `> 0`, `< 0`). Otherwise E0979 with a hint to add the check. Float division is unchanged.
- A shift count must be a constant in `[0, width)` or a name narrowed first (`n >= 0 && n < 64`; an unsigned count needs only the upper bound). Otherwise E0980.
- Remove the runtime division-by-zero and shift-count panics from the emitted C and the runtime.
- Reuse the narrowing of `let` and `is` (let-narrowing). Decide, and record as a D92 amendment: how a narrowing ends when the name is reassigned or passed as `mutable`, and whether a field path (`size.count`) can be narrowed. If field paths cannot, the hint says to copy the field into a local first.

`selfhost/src` and `std/src` must follow the rule in the same commit: fix every site it rejects. Update SPEC.md sections 2, 4 and 6 as needed, and drop "Not built yet" for this part of D92.

## Acceptance

- [ ] Error samples for E0979 and E0980: a variable divisor, a check that does not narrow, a reassigned name, a shift count out of range.
- [ ] Program samples: each narrowing form compiles and runs, for signed and unsigned counts.
- [ ] The emitted C has no division or shift panic path (a dump or C sample shows it).
- [ ] `sh scripts/test.sh` passes.
