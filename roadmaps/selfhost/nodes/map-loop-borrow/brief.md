---
id: map-loop-borrow
title: "Borrow items in for loops over Map keys and entries"
status: todo
priority: should
type: feature
story:
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
spent:
owns:
  - roadmaps/selfhost/nodes/map-loop-borrow/**
shared:
  - selfhost/src/emitmaps.nra
reserves: []
provides: []
depends_on:
  - on: batch-4
    kind: hard
    why: builds on the batch-4 compiler and its measurements
---

# Borrow items in for loops over Map keys and entries

## Goal

`emitMapFor` in `selfhost/src/emitmaps.nra` retains and releases each item of a `for` over `Map.keys` and `Map.entries`. The loop holds the map, so items can be borrowed the same way perceus did for arrays.

## Context

- Follow the perceus rule for array items (`git show feat/batch-4:roadmaps/selfhost/nodes/perceus/outcome.md`): the loop must hold the map for the whole loop, including when the body reassigns or mutates it.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [ ] Samples where the body reassigns, inserts into and removes from the map it iterates, clean under AddressSanitizer; a mutant compiler without the loop hold fails them. Instructions for `narrac emit selfhost` before and after.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
