---
id: wrapping-arithmetic
title: "Integer overflow wraps in every build"
status: todo
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
  - testdata/wrapping*
  - testdata/errors/wrapping*
  - roadmaps/selfhost/nodes/wrapping-arithmetic/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [code E0971]
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 compiler
---

# Integer overflow wraps in every build

## Goal

D92: `+`, `-`, `*` and unary `-` wrap in every build, so the debug overflow panics go; `MIN / -1` gives `MIN` and `MIN % -1` gives `0`. Remove `+%`, `-%`, `*%` and `+%=`, `-%=`, `*%=`: using one is error E0971 with a fix that writes the plain operator; replace the 28 uses in `selfhost/src`, `std/src` and the tests. Add checked functions returning `Option`: `Int.addChecked`, `subChecked`, `mulChecked`, and the same in `std/i8` to `std/u64`. Division by zero and shift counts are not part of this node (proven-operands).

## Acceptance

- [ ] Samples: overflow on `int` and each sized type wraps in debug and release builds; `MIN / -1`; the checked functions return `None` on overflow.
- [ ] Error tests for each removed operator, with the fix (E0971).
- [ ] The panic samples that expected an overflow panic are updated or removed.
- [ ] SPEC.md sections on arithmetic and D76's compound forms updated; D92 drops "Not built yet" for this part.
- [ ] `sh scripts/test.sh` passes.
