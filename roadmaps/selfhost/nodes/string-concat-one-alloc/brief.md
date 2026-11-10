---
id: string-concat-one-alloc
title: "Build string + chains in one allocation"
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/string-concat-one-alloc/**
shared:
  - selfhost/src/emitexprs.nra
reserves: []
provides: []
depends_on:
  - on: batch-4
    kind: hard
    why: builds on the batch-4 compiler and its measurements
---

# Build string + chains in one allocation

## Goal

template-one-alloc builds template literals with one `lstr_concat_n` call, but `a + b + c` chains and `+=` still call `lstr_concat` once per operator. Emit a flat chain as one `lstr_concat_n` call.

## Context

- Read `git show feat/batch-4:roadmaps/selfhost/nodes/template-one-alloc/outcome.md`, including the evaluation-order fix: earlier operands are held when a later one may write.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [ ] A sample with chains of 2, 3 and many operands, including an operand that assigns a variable read by an earlier one, with output captured from the base compiler. Instructions for `narrac emit selfhost` and `bench/strings` before and after.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.
