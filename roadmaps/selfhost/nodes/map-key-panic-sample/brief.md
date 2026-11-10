---
id: map-key-panic-sample
title: "Panic sample for a compound assignment on a missing map key"
status: done
priority: should
type: chore
story: batch-4
early: false
outline: false
milestone: m6-performance
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
added: 2026-11-10
spent: 48120
owns:
  - testdata/panics/map_missing_key*
  - roadmaps/selfhost/nodes/map-key-panic-sample/**
shared:
  - selfhost/tests/src/panics.nra
  - roadmaps/selfhost/inbox.md
reserves: []
provides: []
depends_on: []
---

# Panic sample for a compound assignment on a missing map key

## Goal

`counts[key] += 1` on a missing `Map` key panics, but no panic sample covers it. Add a `PanicSample` in `selfhost/tests/src/panics.nra` with the message "map key not found" (check the exact runtime text) and its program next to the other panic samples.

## Acceptance

- [ ] The panics suite runs the new sample and matches the message.
- [ ] Inbox entry removed.

## Sizing

S: one sample and one table row.
