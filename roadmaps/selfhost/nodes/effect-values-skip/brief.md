---
id: effect-values-skip
title: "The effect pass fills valuesAt only for bodies that read it"
status: done
priority: should
type: feature
story: batch-5
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
spent: 78700
owns:
  - roadmaps/selfhost/nodes/effect-values-skip/**
shared:
  - selfhost/src/effects.nra
reserves: []
provides: []
depends_on:
  - on: profile-compiler-2
    kind: hard
    why: the profile is the evidence for this change
---

# The effect pass fills valuesAt only for bodies that read it

## Goal

`walkExpression` (`selfhost/src/effects.nra:947`) records every expression under `packPosition(body, pos)` in the `valuesAt` map. Its only reader, `findStoredValue` (`effects.nra:323`), looks it up by `FunctionStore.pos`. Skip filling `valuesAt` for a body that has no `FunctionStore`.

## Context

- Evidence: the `valuesAt` map lookup 1.8% and set 0.6% in the split build. Saves up to 2%, depending on how many bodies hold a function store. Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 6.
- Keying the map by node id is the other option, but `recordFunctionStore` (`selfhost/src/exprs.nra:441` and `:447`) has no node id today.

## Acceptance

- [x] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. The gain must exceed the run-to-run spread, or the node closes as not worth it.
- [x] Effect diagnostics and emitted C are unchanged across the samples and error tests.
- [x] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Other batch-5 nodes change the compiler at the same time, so measure the base again at your own starting commit.
- Do not regenerate the seed.
