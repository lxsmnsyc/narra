---
id: batch-15
title: "Batch 15: the feature review nodes and inbox fixes"
status: active
priority: must
type: story
story: 
early: false
outline: false
milestone: p3-review
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
gate: false
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/batch-15/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-13
    kind: hard
    why: builds the D119 outcomes on the batch-13 seed
---
# Batch 15: the feature review nodes and inbox fixes

## Goal

Fifteenth one-shot batch, scope set by the maintainer on 2026-11-14: the five accepted D119 items (field shorthand, `Co.send`, removing `lazy`, the `@inline` benchmark, `Ranges.stepped`), the inbox correctness bugs, the generator type gaps, small review nits, and const-functions-std. Children land on `feat/batch-15` as direct commits after a worker check and a fresh review. At the end: seed updates, field-shorthand-sweep, full check and Linux check, debug copy check, one fresh review of the whole branch, bench row, one pull request.

Children:

- p3 review: field-shorthand, coroutine-send, remove-lazy, inline-benchmark (spike, gate on any removal), remove-inline (added after the maintainer accepted the benchmark's result, 2026-11-14), field-shorthand-sweep (last, after the seed update).
- p2 std: range-step, const-functions-std.
- Fixes: float-shortest, slot-shared-writes, index-assign-order, stream-exit (gate: amends D109 item 6), generator-option-params, batch-15-nits.

## Order and shared files

- Machine: 8 GB, at most four building workers at once; hold new builds while load is above 30. inline-benchmark runs alone at the end of the batch on a quiet machine, before the bench row.
- Most nodes touch `selfhost/src/*.nra`; feature logic goes in its own module where it can, shared files carry dispatch rows only. The second to land rebases and keeps both sides.
- `diagnostics.nra` is sorted by code and guarded by the `codes` suite: insert rows at their code position, never append.
- Codes: range-step E1270-E1279, field-shorthand E1280-E1284. remove-lazy retires E1032-E1035.
- Generated files (`stdtext.nra`, `runtime*.nra`) only through `sh scripts/test.sh --suite generated --update`; after each landing, run the generated and guide suites.
- Each worker uses a scratch folder named after its node under /private/tmp/claude-501/. Check a rebase's exit code before testing.

## Seed

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. No child uses its own feature in the compiler or std, except field-shorthand-sweep after the batch seed update. Runtime text edits need two seed updates at the end of the batch.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [x] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [x] A debug-built compiler runs `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0.
- [x] A fresh review of the whole story branch, before the pull request.
- [x] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise by more than the noise.
