---
id: map-helpers-split
title: "Map helpers are emitted without Str.replace"
status: todo
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
spent:
owns:
  - roadmaps/selfhost/nodes/map-helpers-split/**
shared:
  - selfhost/src/emitmaps.nra
reserves: []
provides: []
depends_on:
  - on: profile-compiler-2
    kind: hard
    why: the profile is the evidence for this change
---

# Map helpers are emitted without Str.replace

## Goal

`emitmaps.mapFunctions` fills a template with `Str.replace` for each map type. Split the template once into parts and join them with the type names.

## Context

- Evidence: `narra_str_replace` 2.0% inclusive, all from `mapFunctions`. Saves about 2%. Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 8.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. The gain must exceed the run-to-run spread, or the node closes as not worth it.
- [ ] Emitted C is byte-identical for `selfhost` and every sample.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Other batch-5 nodes change the compiler at the same time, so measure the base again at your own starting commit.
- Do not regenerate the seed.
