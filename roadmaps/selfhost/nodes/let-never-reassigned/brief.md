---
id: let-never-reassigned
title: "A `let` that is never reassigned is a compile error"
status: done
priority: could
type: feature
story: batch-16
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
added: 2026-11-14
spent: 271940
owns:
  - roadmaps/selfhost/nodes/let-never-reassigned/**
  - testdata/let_const*
  - testdata/errors/let_const*
shared:
  - selfhost/src/*.nra
  - selfhost/src/diagnostics.nra
  - std/src/*.nra
  - testdata/**
  - SPEC.md
  - compiler/README.md
  - docs/guide/*.md
reserves: [diag E1510-E1519]
provides: []
depends_on:
  - on: dead-code-check
    kind: hard
    why: both sweep most of selfhost/src and testdata; landing them in order avoids a wide rebase
---
# A `let` that is never reassigned is a compile error

## Goal

A `let` binding that is never assigned after its declaration is an error with a fix that rewrites it to `const`. This replaces the "should be const" rule of the parked lint node (see lint, dropped).

## Context

- `selfhost/src` has about 2,100 `let` lines. Measure how many the rule flags before writing code, and report the count in `log.md`. If the sweep is far over 500 sites, stop and ask the orchestrator whether to land the rule first and sweep in a second node.
- What counts as a write: `x = ...`, compound assignment, `x[i] = ...`, `x.f = ...`, and passing `mutable x`. Writes inside closures count.
- Severity: error, like unused names (open question in batch-16).
- Seed rule: the compiler source must compile with the committed seed. Do the sweep with the batch's built compiler through `narrac apply-fix`, and keep the sweep in its own commit so it can be reviewed apart from the rule.

## Acceptance

- [ ] E1510 on every `let` never written after its declaration, with an `apply-fix` edit to `const`.
- [ ] Selfhost, std, samples and guide samples pass under the rule.
- [ ] SPEC.md states the rule next to "Unused names are errors"; the guide chapter on values teaches it.
- [ ] Error tests in `testdata/errors/let_const*`.

## Sizing

M: a write set per function plus a wide mechanical sweep.
