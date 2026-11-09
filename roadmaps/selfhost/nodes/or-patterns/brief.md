---
id: or-patterns
title: "Or-patterns in match arms"
status: active
priority: should
type: feature
story: batch-3
early: false
outline: false
milestone: m12-spec-gaps
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
added: 2026-11-09
spent:
owns:
  - testdata/or_pattern*
  - testdata/errors/or_pattern*
  - testdata/fmt/*or_pattern*
  - roadmaps/selfhost/nodes/or-patterns/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [code E0967, code E0968, code E0974]
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 compiler
---

# Or-patterns in match arms

## Goal

Build D98: alternatives with `|` in `match` arm patterns, nested too (`Some(1 | 2)`). Every alternative binds the same names with the same types; a guard covers the whole or-pattern; a whole-value binding keeps the full type (no narrowing, no union types, D73); exhaustiveness and unreachable-arm checks count each alternative. Not allowed in `is` (a clear error with a hint to use `match`). `narrac fmt`: one line when it fits, otherwise one alternative per line starting with `|`.

New AST node kind: checker, emitter, `fmt` and `astdump` cases. Do not use or-patterns in `selfhost/src` (adopt-new-syntax does, after the mid-batch seed).

## Acceptance

- [ ] Samples `testdata/or_pattern*.nra` with `.out`, clean under AddressSanitizer: constructor alternatives, literal alternatives, nested `Some(1 | 2)`, alternatives that bind the same name, a guard on an or-pattern.
- [ ] Error tests (E0967, E0968, E0974): alternatives binding different names, the same name with different types, an or-pattern in `is`; an alternative already covered is reported unreachable.
- [ ] fmt tests for the short and long layouts.
- [ ] SPEC.md describes or-patterns; D98 loses "Not built yet".
- [ ] `sh scripts/test.sh` passes.
