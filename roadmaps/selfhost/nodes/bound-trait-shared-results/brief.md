---
id: bound-trait-shared-results
title: "Shared results through a bound <T: Trait> count as fresh"
status: active
priority: should
type: fix
story: batch-12
early: false
outline: false
milestone: 
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
added: 2026-11-13
spent:
owns:
  - testdata/errors/shared_result_bound_*
  - roadmaps/selfhost/nodes/bound-trait-shared-results/**
shared:
  - selfhost/src/sharedresults.nra
reserves: []
provides: []
depends_on: []
---
# Shared results through a bound <T: Trait> count as fresh

## Goal

A call through a bound `<T: Items>` to a trait function declared `items(value: Self): Array<int> from value` counts as a fresh result in `findSharedArgument` (`selfhost/src/sharedresults.nra` around line 140), so `let got = Items.items(item); Arrays.push(mutable got, 3)` compiles and the debug copy check exits 103. Read the `from` marker of the trait function for such calls.

## Context

- From the dynamic-traits review, 2026-11-12. On the base too.
- Proof first: the program compiles and exits 103 under `NARRA_CHECK_COPIES=1` on the base.

## Acceptance

- [ ] Error test for the bound call; the same code with `copy(...)` compiles and runs clean.
