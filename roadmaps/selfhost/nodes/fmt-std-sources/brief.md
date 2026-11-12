---
id: fmt-std-sources
title: "The sources suite checks std/src, and fmt levels block operands in a chain"
status: todo
priority: should
type: fix
story: batch-10
early: false
outline: false
milestone: 
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
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/fmt-std-sources/**
  - selfhost/tests/src/sources.nra
  - selfhost/src/fmtexprs.nra
  - std/src/**
shared:
  - selfhost/src/stdtext.nra
  - testdata/fmt/**
reserves: []
provides: []
depends_on: []
---

# The sources suite checks std/src, and fmt levels block operands in a chain

## Goal

Two findings from fmt-block-hug (batch-9):

1. `std/src/coroutine.nra` and `std/src/option.nra` are not formatted, because the `sources` suite (`selfhost/tests/src/sources.nra`) does not check `std/src`. Extend the suite to `std/src` and format every file it flags. `selfhost/src/stdtext.nra` is generated from `std/src`: rewrite it with `sh scripts/test.sh --suite generated --update`.
2. When a block is not the last operand of an assignment chain (`const mid = x + {...} + y`), oxfmt prints all operands level, but `narrac fmt` keeps the extra indent, because `endsWithBlock` (`selfhost/src/fmtexprs.nra`, near line 598) checks only the rightmost operand. Match oxfmt (see `selfhost/CLAUDE.md`), and add a `fmt` golden for the case.

Do not regenerate the seed; the orchestrator does that at the end of the batch.

## Acceptance

- Proof on `feat/batch-10`: the `fmt --check` result for `std/src`, and the current output for the chain case, in `log.md`.
- The `sources` suite covers `std/src` and passes.
- A `fmt` golden for a block in the middle of an assignment chain matches oxfmt's layout.
- `sh scripts/test.sh` passes.
