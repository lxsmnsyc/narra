---
id: quote-on-error
title: "Quote local names only when reporting E0601"
status: active
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
spent:
owns:
  - roadmaps/selfhost/nodes/quote-on-error/**
shared:
  - selfhost/src/funcvalues.nra
reserves: []
provides: []
depends_on:
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# Quote local names only when reporting E0601

## Goal

`useLocal` (`selfhost/src/funcvalues.nra:73`) calls `quoteGo(local.name)` on every use of a local but needs the quoted name only when it reports E0601. Quote it inside the error branch.

## Context

- Evidence: `parser.quoteGo` 2.3% inclusive, 99% of it from `useLocal`. Profile proposal 10, saves about 2%. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] The E0601 error tests print the same message.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

S: one file, a measured hotspot.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
