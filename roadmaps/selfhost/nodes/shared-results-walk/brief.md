---
id: shared-results-walk
title: "A cheaper walk for the from check"
status: todo
priority: should
type: refactor
story: batch-7
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/shared-results-walk/**
shared:
  - selfhost/src/sharedresults.nra
  - selfhost/src/body.nra
  - selfhost/src/check.nra
reserves: []
provides: []
depends_on: []
---

# A cheaper walk for the from check

## Goal

The `from` check (E0976, `checkSharedResults` in `selfhost/src/sharedresults.nra`) costs about +3% instructions on `narrac emit selfhost`, even with the `Checker.heldTypeSets` cache (inbox, from shared-write-sweep-rest). Profile it and make the walk cheaper. Keep the signatures of `findSharedArgument` and `findExpressionStart`, which `sharedwrites.nra` imports.

## Context

- Profile and C sizes from the batch-7 planning session (2026-11-11, commit 94360a39, Apple M2): `narrac emit selfhost` is 2.49G instructions and 138 MB peak.
- Profile with `sample` on a `-O2 -g` build of the emitted compiler: `sample narrac 2 1 -wait -mayDie -file out.txt` started just before `narrac emit selfhost` (the planning session did this).
- If the profile shows the cost is not worth a change (under 2% and no clear hot spot), record the numbers in the outcome and close the node with no code change.

## Acceptance

- [ ] Every `from`/E0976 sample gives the same diagnostics as before.
- [ ] The outcome records the profile and the instructions of `emit selfhost` before and after (median of 3).
- [ ] Fast check passes: `sh scripts/test.sh`
