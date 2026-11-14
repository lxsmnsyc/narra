---
id: batch-13
title: "Batch 13: finish the syntax (phase p1)"
status: done
priority: must
type: story
story: 
early: false
outline: false
milestone: p1-syntax
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/batch-13/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-12
    kind: hard
    why: p1 builds on the async runtime and the batch-12 seed
---
# Batch 13: finish the syntax (phase p1)

## Goal

Thirteenth one-shot batch, scope set by the maintainer on 2026-11-13: finish every syntax node of phase p1, so the language is final for the seed compiler, then run the feature review (p3) on the finished language. Runs beside batch-14 (std, p2) on its own branch. Children land on `feat/batch-13` as direct commits after a worker check and a fresh review. At the end: seed updates, full check and Linux check, debug copy check, one fresh review of the whole branch, bench row, one pull request.

Regions are dropped by D113 (maintainer, 2026-11-13): re-measured, they no longer pay off after the size-class allocator.

Children:

- Design spikes, each with a maintainer gate: async-generators-design (D110), threads-design (D111), simd-design (D112), regions-plan (D113), tagged-templates-design (D114), parameter-defaults-design (D115), refinement-types-design (D116), uses-list-design (D117).
- Build nodes, refined by their spikes: async-generators, threads (split by its spike), simd, tagged-templates, parameter-defaults, refinement-types, uses-list-syntax.
- feature-review (p3, last): a goal debate over the finished language. It ends with proposals; accepted ones become nodes of a later batch.

A design that cannot be settled at its gate leaves the batch with a reason here; the batch ships without it.

## Order and shared files

- Machine: 8 GB, four building workers at a time. Spikes build little; run up to four beside the builders.
- All spikes start at once; the maintainer answers the gates together. Builds start as their gates are answered.
- Parser, `ast.nra`, `astdump.nra`, `fmt.nra` and `diagnostics.nra` are shared by most builds: feature logic goes in its own module, shared files carry only dispatch rows; the second to land rebases and keeps both sides.
- Decisions D110 to D117 are reserved per spike. Codes: async-generators E1112-E1129 (left free in the async-core range), threads E1190-E1199 and E1400-E1419, simd E1250-E1269, tagged templates E1200-E1209, parameter defaults E1210-E1219, refinement types E1220-E1239, uses lists E1240-E1249.
- batch-14 edits std sources and the generated `stdtext.nra` on its own branch. Whichever batch merges second rebases, regenerates generated files and the seed.
- Each worker uses a scratch folder named after its node under /private/tmp/claude-501/. Check a rebase's exit code before testing.

## Seed

`selfhost/src` and `std/src` may use only features the committed seed compiles. No child uses its own feature in the compiler or seed-compiled std. Runtime text edits need two seed updates at the end of the batch.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [x] A debug-built compiler runs `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0.
- [x] SPEC has no "not built yet" or "not in the first version" note left for a p1 feature.
- [x] feature-review's proposals are handed to the maintainer.
- [x] A fresh review of the whole story branch, before the pull request.
- [x] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise by more than the noise.
