---
id: adopt-new-syntax
title: "Use compound assignment, C-style for and or-patterns in the compiler"
status: active
priority: should
type: refactor
story: batch-3
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - roadmaps/selfhost/nodes/adopt-new-syntax/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: or-patterns
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
  - on: missing-return
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
  - on: wrapping-arithmetic
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
  - on: type-aliases
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
  - on: decls-quadratic
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
  - on: template-positions
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
  - on: lsp-navigation-parse
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
  - on: unused-lines
    kind: hard
    why: the sweep touches every hot file and uses the new syntax, so it runs last, after the mid-batch seed
---

# Use compound assignment, C-style for and or-patterns in the compiler

## Goal

The compiler's own code adopts the syntax batch-2 and batch-3 added (maintainer, 2026-11-09):

- `x = x + 1` and `x = x - 1` become `x++` and `x--` (404 lines measured); `x = x op y` becomes `x op= y` (457 lines).
- A `while` loop becomes a C-style `for` only when its counter is set by the statement right before it, its step is the loop body's last statement, the counter is not used after the loop, and the body has no `continue` (where `continue` would now run the step and change the meaning). Leave every other `while` as it is.
- Neighbouring `match` arms with identical bodies merge into an or-pattern (177 pairs measured).

Scope: `selfhost/src`, `std/src` (edit in place; std line numbers appear in sample C, so check the emit goldens), `selfhost/tests/src`. No behavior change: every sample's output and the bootstrap fixed point hold. One commit per group of files. Requires the mid-batch seed (the orchestrator regenerates it before this node starts).

## Acceptance

- [ ] Counts before and after for each rewrite in the outcome, and the loops left as `while` on purpose with the reason.
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point; every sample's output is unchanged.
- [ ] `sh scripts/test.sh` passes, including the `sources` format check.
