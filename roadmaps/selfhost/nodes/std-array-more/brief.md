---
id: std-array-more
title: "Array sorting and the missing array functions"
status: active
priority: could
type: feature
story: batch-22
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-array-more/**
shared:
  - selfhost/src/std/**
  - selfhost/runtime/**
  - selfhost/src/runtime/**
  - std/src/**
  - testdata/**
  - selfhost/tests/src/**
  - docs/**
  - scripts/**
  - SPEC.md
reserves: []
provides: []
depends_on: []
---
# Array sorting and the missing array functions

## Goal

Survey items 1 and 3: `Arrays.sort` and `Arrays.sortBy` (stable; `sort` over `Ordered` values), `pop`, `truncate` (runtime builtins that shrink in place), `contains`, `indexOf`, `reverse`. Chapter 11's folder sample can then print sorted names instead of a count (the inbox entry about sorting closes).

## Rules for every std addition

- Read `roadmaps/selfhost/nodes/std-gap-survey/survey.md` for the evidence of each piece.
- Write it in Narra in `std/src` where it can be; use the runtime only where the survey says so.
- Each export gets a registry row (or source declaration) with a readable display signature (see std-registry-signatures' outcome); the `query/builtin_signatures` case must pass.
- Replace the hand-written copies in `selfhost/src`, `scripts/` and `testdata/` that the survey lists, where the seed rule allows (`selfhost/src` may use only what the committed seed compiles: a new std function is usable in `selfhost/src` only after the seed is regenerated, so leave compiler call sites for a follow-up and list them in the outcome).
- Teach it in the guide where the survey's evidence points (one short runnable sample each) and regenerate the std reference (`sh scripts/test.sh --suite generated --update`).

## Acceptance

- [ ] Each function exists with a case per function (and the edge cases the survey names), a display signature, a guide sample, and a regenerated std reference page.
- [ ] `sh scripts/test.sh` passes; `pnpm --dir website build` passes.
