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

- Spikes, wave 1: profile-compiler-2 (profile `narrac emit selfhost` again), profile-programs (sieve and particles against Go and Rust) and interp-selects (the interp cost of inline enum payloads, D99).
- Follow-ups from batch-4: cache-type-names-rest, string-concat-one-alloc, map-loop-borrow and tail-loop-borrow. Check each against the profile-compiler-2 numbers before it starts, as that brief asks.
- D94 steps 1 and 2: copy-builtin, then from-marker. They run one after the other, as the debate orders them (P20), because both edit the checker's sharing facts. The compiler does not use them until shared-write-sweep.
- The spikes may propose more children; the orchestrator adds the ones the maintainer accepts.

Benchmark children hold the `benchmarks` resource, so they never run at the same time.

## Seed

The performance children change the emitted C. Regenerate the seed after the follow-ups land and again at the end.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `bench/BASELINE.md` has a row for the start of the batch and one for the end, recorded on an idle machine, including the compiler row.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [ ] The seed is regenerated at the end and `sh selfhost/bootstrap.sh` reports it current.
