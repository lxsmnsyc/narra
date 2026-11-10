---
id: batch-5
title: "Batch 5: performance follow-ups and value-semantics markers"
status: active
priority: should
type: story
story:
early: false
outline: false
milestone: m6-performance
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/batch-5/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 5: performance follow-ups and value-semantics markers

## Goal

Fifth one-shot batch, scope set by the maintainer on 2026-11-10: the performance follow-ups planned after batch-4 (commit 803fa7c9), plus the first two D94 nodes. Children land on `feat/batch-5` as direct commits after targeted tests. At the end the batch gets one full check, one Linux check, one seed update, one fresh review and one pull request.

Children:

- Spikes, done: profile-compiler-2, profile-programs and interp-selects (no node: D99's cost stands).
- From profile-programs: array-write-cold-paths, then loop-unique-hoist (excludes loops in `function*`).
- From profile-compiler-2, ranked: tail-loop-borrow first, then cli-skip-teardown, lexer-punctuation, body-id-ranges, scope-map, effect-values-skip, cache-type-names-rest (widened to declaration names and module hashes), map-helpers-split, and parser-token-kinds after lexer-punctuation.
- D94 steps 1 and 2: copy-builtin, then from-marker. They run one after the other, as the debate orders them (P20), because both edit the checker's sharing facts. The compiler does not use them until shared-write-sweep.
- Dropped on 2026-11-10 with the spike evidence: string-concat-one-alloc and map-loop-borrow.

Benchmark children hold the `benchmarks` resource, so they never run at the same time.

## Seed

The performance children change the emitted C. array-write-cold-paths edits the runtime and regenerates the seed twice on its branch. Regenerate it again after the compiler-speed children land and at the end.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `bench/BASELINE.md` has a row for the start of the batch (the end-of-batch-4 row at afb1eac3 serves: no compiler, runtime or seed change landed between it and the batch-5 base) and one for the end, recorded on an idle machine, including the compiler row.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [ ] The seed is regenerated at the end and `sh selfhost/bootstrap.sh` reports it current.
