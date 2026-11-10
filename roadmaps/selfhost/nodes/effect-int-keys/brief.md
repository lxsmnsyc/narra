---
id: effect-int-keys
title: "Integer keys in the effect pass"
status: done
priority: should
type: fix
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
spent: 86944
owns:
  - roadmaps/selfhost/nodes/effect-int-keys/**
shared:
  - selfhost/src/effects.nra
reserves: []
provides: []
depends_on:
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# Integer keys in the effect pass

## Goal

`selfhost/src/effects.nra` keys `valuesAt`, `bodyIndex`, `entries`, `sources` and `locals` by integers (body index and node id, or a packed position) instead of the strings built by `positionKey`, `functionKey` and `localKey`. Also stop copying the `valuesAt` array on every push: `walkExpression` reads the array from the map, pushes and stores it back while the map still holds it, so copy-on-write copies it each time.

## Context

- Evidence: `positionKey` 6.1%, `functionKey` 2.5% and `localKey` 0.6% inclusive, out of 15.9% for the effects phase. Profile proposal 2, saves up to 8%. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] Every effect diagnostic and `narra query` effect result is unchanged (the existing error and query suites cover them).
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

S: one file, a measured hotspot.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
