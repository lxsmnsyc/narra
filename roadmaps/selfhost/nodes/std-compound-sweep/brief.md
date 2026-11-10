---
id: std-compound-sweep
title: "Compound assignment in std sources"
status: done
priority: should
type: chore
story: batch-6
early: false
outline: false
milestone: m10-std
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/std-compound-sweep/**
shared:
  - std/src/**
  - selfhost/src/stdtext.nra
  - testdata/**
reserves: []
provides: []
depends_on:
  - on: std-no-panic
    kind: hard
    why: both edit std/src; std-no-panic changes the same functions
---

# Compound assignment in std sources

## Goal

From `roadmaps/selfhost/inbox.md` (2026-11-10, from adopt-new-syntax): `std/src` still writes `x = x + 1` (28 sites) and `x = x op y` (5). Rewrite them as compound assignment and accept the C change it brings: 8 std samples change their emitted C (a pointer temporary, renumbered temporaries, a division check that moves column). Regenerate with `sh scripts/test.sh --suite generated --update` and check the diffs are only those.

## Acceptance

- [ ] No `x = x op y` left in `std/src` (grep in the outcome).
- [ ] `sh scripts/test.sh` passes.
