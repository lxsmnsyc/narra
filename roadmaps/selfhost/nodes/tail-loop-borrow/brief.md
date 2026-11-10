---
id: tail-loop-borrow
title: "Tail-loop functions borrow the parameters they pass through"
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
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/tail-loop-borrow/**
shared:
  - selfhost/src/emittail.nra
reserves: []
provides: []
depends_on:
  - on: batch-4
    kind: hard
    why: builds on the batch-4 compiler and its measurements
---

# Tail-loop functions borrow the parameters they pass through

## Goal

`openTailLoop` (`selfhost/src/emittail.nra`) retains every non-`mutable` parameter at entry so a tail jump can replace it. A parameter that every tail call passes back unchanged (`emitter` in `needsRc`, `checker` in `collectUses`) needs no retain and no release. The first profile put this at 4.8% of `narrac emit selfhost` across 25 functions; const-params-pointer changed how these parameters are passed, so measure again first.

## Context

- Read `git show feat/batch-4:roadmaps/selfhost/nodes/const-params-pointer/outcome.md` (self tail calls now hold large parameters in `<name>_held`) and the perceus outcome.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [ ] Instructions for `narrac emit selfhost` before and after, 3 runs each, the gain above the spread. A sample with a tail-recursive function that passes a counted parameter unchanged and one that replaces it, clean under AddressSanitizer and the leak check.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
