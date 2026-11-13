---
id: review-nits
title: "Small fixes from the batch-11 reviews"
status: active
priority: should
type: chore
story: batch-12
early: false
outline: false
milestone: 
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - testdata/errors/value_prelude_*
  - roadmaps/selfhost/nodes/review-nits/**
shared:
  - selfhost/src/stdvalues.nra
  - selfhost/src/emitinline.nra
  - selfhost/src/inlining.nra
  - selfhost/src/driver.nra
  - selfhost/src/diagnostics.nra
reserves: []
provides: []
depends_on: []
---
# Small fixes from the batch-11 reviews

## Goal

From the inbox, 2026-11-12:

- `writeArrowHint` (`selfhost/src/stdvalues.nra`) suggests `(items) => ...` for a zero-parameter builtin; write `() => Name()` there.
- Prelude functions (`print`, `assert`, `panic`) as values give `unknown name`; give them the E1013 hint.
- Arrow loop labels (`emitinline.nra` around line 220, `_i<tempCount>`) and copy labels (around line 290, `_i<inlineCount>`) share one prefix with two counters; give arrows their own prefix.
- E1030 walks only instance bodies (`inlining.nra` around lines 118-129), so an exported generic recursive `@inline` function its package never calls compiles clean. Check declared generic `@inline` functions too.
- `runBuild` in `selfhost/src/driver.nra` returns 1 with a full release on its error paths; use `exitWithoutRelease(1)`.

## Context

- async-core edits `emitinline.nra` around the same lines (`inGenerator`); the second to land rebases and keeps both.

## Acceptance

- [ ] Error tests for the hint and for the uncalled generic `@inline` function; a sample with an inlined arrow and a copy in one function.
