---
id: batch-23
title: "Batch 23: the compiler uses the new std, clean std headers"
status: active
priority: should
type: story
story:
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/batch-23/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-22
    kind: hard
    why: the compiler can use the std functions batch-22 added only with its seed
---
# Batch 23: the compiler uses the new std, clean std headers

## Goal

Before the `v0.1.0` tag (maintainer, 2026-11-20: "begin run"):

- The compiler replaces its hand-written sorts, sets, character checks and shell calls with the std functions batch-22 added (compiler-std-sorts, compiler-std-sets, compiler-std-files).
- std header comments stop citing SPEC sections and decision numbers (std-header-comments).

The branch `feat/batch-23` starts from `main` after PR #129. Worktree `../narra-batch-23`, local only.

Outside the batch, already merged or open: C in the benchmarks (PR #129, Odin removed), the Vercel ignore step (PR #130).

## Acceptance

- [ ] Every child done or out of the batch with a reason.
- [ ] `sh scripts/test.sh` passes; the seed is current; `pnpm --dir website build` passes; `sh scripts/linux-check.sh` passes.
