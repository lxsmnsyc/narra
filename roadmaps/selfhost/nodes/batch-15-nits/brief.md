---
id: batch-15-nits
title: "Small fixes from the batch-13 reviews"
status: done
priority: must
type: fix
story: batch-15
early: false
outline: false
milestone: p2-std
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
added: 2026-11-14
spent: 110000
owns:
  - roadmaps/selfhost/nodes/batch-15-nits/**
  - testdata/fmt/uses_list_comments*
  - testdata/errors/simd_load_fixed*
shared:
  - selfhost/src/fmt.nra
  - selfhost/src/stdsimd*.nra
  - selfhost/src/diagnostics.nra
  - docs/guide/12-narra-for-typescript-programmers.md
  - roadmaps/selfhost/inbox.md
reserves: []
provides: []
depends_on: []
---
# Small fixes from the batch-13 reviews

## Goal

- fmt: a comment inside a `uses` list moves. `Generator<int> uses (io, /* c */ alloc)` formats as `Generator<int /* c */> uses (io, alloc)`, and `uses /* c */ io` has the same bug. Keep comments where written (oxfmt behavior). fmt golden fails on the base.
- `Simd.load` on a FixedArray says `expected Array<i32>, found FixedArray<i32, 4>`: add a hint naming `Simd.fromFixed`. Error golden.
- The guide chapter 12 (around line 90) says there is no C-style `for` loop; it compiles. Fix the text.
- Remove the inbox entry about the diagnostics guard (done in batch-13) and the entries this batch fixes; the orchestrator lands the inbox edit.

## Acceptance

- [ ] Goldens fail on the base and pass. `sh scripts/test.sh` passes.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
