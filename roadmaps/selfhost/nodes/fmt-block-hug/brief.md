---
id: fmt-block-hug
title: "narrac fmt hugs a block expression in the last argument and as an operand"
status: todo
priority: should
type: fix
story: batch-9
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/fmt-block-hug/**
  - selfhost/src/fmtexprs.nra
  - selfhost/src/fmtdoc.nra
  - selfhost/src/fmtprinter.nra
shared:
  - selfhost/src/fmt.nra
  - testdata/**
  - selfhost/src/**
  - std/src/**
reserves: []
provides: []
depends_on: []
---

# narrac fmt hugs a block expression in the last argument and as an operand

## Goal

Inbox, from chore/fmt-sources: `narrac fmt` does not hug a block expression in the last call argument. `addAll(mutable xs, { ... })` becomes one argument per line with the block on its own line, where oxfmt keeps a last-argument function or object on the call line. A block as a binary operand (`x + { return f() }`) breaks into `x +` and an indented block on the next lines. Match oxfmt (the project follows oxfmt, not Prettier, for layout). See `compiler/testdata/errors/block_expressions_mutable_arg.nra` and `mutable_read_conflict_block_operand.nra`.

The `sources` suite checks that the compiler and std are formatted, so reformat any file the new layout changes in the same commit; touch nothing else in those files.

## Acceptance

- fmt cases for both layouts, plus a case where the hug does not fit and falls back.
- `narrac fmt` is idempotent on the cases and on the reformatted sources.
- `sh scripts/test.sh` passes.
