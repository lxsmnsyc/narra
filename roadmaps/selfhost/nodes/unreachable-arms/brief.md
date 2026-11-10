---
id: unreachable-arms
title: "Unreachable match arms are errors"
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
  - testdata/errors/unreachable_arm_*
  - testdata/diagnostics/unreachable_arm_*
  - roadmaps/selfhost/nodes/unreachable-arms/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
reserves: [diag E0983]
provides: []
depends_on: []
---

# Unreachable match arms are errors

## Goal

From `roadmaps/selfhost/inbox.md` (2026-11-09, from or-patterns): the unreachable check (E0974) runs only on or-pattern arms, so an unreachable plain `match` arm (after a wildcard, or a constructor already covered without a guard) is accepted. Report every unreachable arm with its own code, E0983, and a fix that removes the arm.

Run it over `selfhost/src`, `std/src` and `testdata` first, and fix every arm it rejects in the same commit. List the count in the outcome.

## Acceptance

- [ ] Error samples: an arm after a wildcard, a repeated constructor, a repeated literal. A guarded arm before the same pattern stays allowed.
- [ ] The fix is applied by `narrac apply-fix` (habits suite).
- [ ] `sh scripts/test.sh` passes.
