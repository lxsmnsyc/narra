---
id: batch-8
title: "Batch 8: fixed arrays and C interop"
status: todo
priority: should
type: story
story: 
early: false
outline: false
milestone: m12-spec-gaps
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/batch-8/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-7
    kind: hard
    why: builds on the proven-operands and contracts prover, tree shaking, @target and the seed regenerated at the end of batch-7
---

# Batch 8: fixed arrays and C interop

## Goal

Eighth one-shot batch, scope set by the maintainer on 2026-11-11: two D95 syntax features. Children land on `feat/batch-8` as direct commits after a worker check and a fresh review. At the end: one seed update, the full check and the Linux check, one fresh review of the whole branch, a bench row, and one pull request.

Children:

- Fixed arrays (D95): const-generics, then fixed-arrays.
- C interop step 1 (D95): attributes (a general attribute list), unsafe-ptr, then c-interop (extern blocks), then c-audit-target. Step 2 (`@layout`, C-callable functions) stays with c-abi-export.

## Order and shared files

- Wave 1: const-generics beside attributes. Wave 2: fixed-arrays beside unsafe-ptr (both after const-generics, which owns the Type key and C-name plumbing). Then c-interop, then c-audit-target.
- fixed-arrays and unsafe-ptr each add a `Type` variant to the same exhaustive matches; the second to land resolves those cases and reruns the full check. Emitter code goes in each feature's own module (`emitfixed.nra`, `emitpointers.nra`, `emitexterns.nra`).
- Both chains edit the parser, checker, emitter, `fmt` and the diagnostics table. Each feature puts its logic in modules of its own (for example `constgenerics.nra`, `fixedarrays.nra`, `unsafecode.nra`, `externs.nra`) and touches shared switch files only for dispatch cases, codes and text. A branch that lands second rebases and keeps both sides.
- Lessons from batch-7 (wave-20): every worker uses a scratch folder named after its node under /private/tmp/claude-501/; a targeted suite that finishes in seconds may have run no cases (inbox), so trust only the full run; check a rebase's exit code before testing.

## Seed

The compiler's source must not use the new syntax. Regenerate the seed once at the end.

## Acceptance

- [ ] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [ ] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [ ] A fresh review of the whole story branch, before the pull request.
- [ ] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions on the batch-7 source do not rise by more than 2%.
