---
id: body-id-ranges
title: "The parser records each body's node id range"
status: dropped
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
spent: 90369
owns:
  - roadmaps/selfhost/nodes/body-id-ranges/**
shared:
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/check.nra
  - selfhost/src/idcheck.nra
reserves: []
provides: []
depends_on:
  - on: profile-compiler-2
    kind: hard
    why: the profile is the evidence for this change
---

# The parser records each body's node id range

## Goal

The parser records the first and last node id of each function, block and constant body, so `check.startBody` (`selfhost/src/check.nra:1540`) does not walk the body with `idcheck.collectFuncIds` to find them. `narrac pkg --ids` keeps using the walker. Parser ranges may be a superset with gaps, which is fine; deriveshape and the prelude parse their own ranges and must keep working.

## Context

- Evidence: `startBody` 3.5% inclusive, of which `collectFuncIds` 2.8%. Saves about 3%. Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 4.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. The gain must exceed the run-to-run spread, or the node closes as not worth it.
- [ ] Emitted C for `selfhost` is byte-identical.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Other batch-5 nodes change the compiler at the same time, so measure the base again at your own starting commit.
- Do not regenerate the seed.
