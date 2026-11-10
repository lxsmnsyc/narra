---
id: batch-4
title: "Batch 4: the performance batch"
status: todo
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
  - roadmaps/selfhost/nodes/batch-4/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 4: the performance batch

## Goal

Fourth one-shot batch, scope chosen by the maintainer on 2026-11-10: make the self-hosted compiler and the programs it builds faster by spending less on reference counts and allocation. Children land on `feat/batch-4` as direct commits after targeted tests; at the end one full check, one Linux check, one fresh review and one pull request.

Children:

- Wave 1, in parallel: bench-compiler (the baseline every perf child measures against), regions-design (spike, no code), and the small fixes from the inbox: closure-arg-temps, template-escapes, nested-or-guards, map-key-panic-sample, typed-calls-and-effects.
- Then profile-compiler (spike), and perceus after the two emitter fixes.
- Then option-inline, then perceus-reuse and const-params-pointer.
- profile-compiler may propose more children; the orchestrator adds the ones the maintainer accepts.

Benchmark children hold the `benchmarks` resource, so they never run at the same time.

## Seed

Every child changes the emitted C. The seed is regenerated after perceus, after option-inline and at the end, so later children and the compiler itself run on the new code.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `bench/BASELINE.md` has a row for the start of the batch and one for the end, recorded on an idle machine, including the compiler row.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [ ] The seed is regenerated at the end and `sh selfhost/bootstrap.sh` reports it current.
