---
id: std-no-panic
title: "Std functions return Result or Option instead of panicking"
status: todo
priority: should
type: feature
story: batch-6
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
  - testdata/std_no_panic_*
  - roadmaps/selfhost/nodes/std-no-panic/**
shared:
  - selfhost/src/**
  - selfhost/runtime/**
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
  - on: batch-2
    kind: hard
    why: builds on the compound assignment, C-style for and std modules batch-2 added
---

# Std functions return Result or Option instead of panicking

## Goal

Build the std part of D92 (DECISIONS.md):

- `Str.slice`, `Str.split`, `Str.repeat`, `Bytes.at`, `Bytes.slice`, `Float.truncate`, `Arrays.filled`, `Ranges.stepped`, `Math.clamp`, `Math.powInt` and `Random.nextInt` stop panicking on bad input. Use `Option` where there is one failure and `Result` where the caller needs the reason. List each in a table in SPEC.md.
- Remove compound assignment on a Map index target (`counts[key] += 1`). Per the no-deprecation rule it gets no special error. Add `Maps.update(mutable map, key, initial, change)`.
- Update every caller in `selfhost/src`, `std/src`, the test runner, the guide and the samples.
- Drop "Not built yet" for this part of D92.

Some of these are builtins in `selfhost/src/std*.nra`, some are source modules in `std/src`. A builtin whose signature changes needs the seed rule: `selfhost/src` may only call the new form after the seed reads it. If the compiler calls one of them, check whether the seed's builtin table already has the new form; if not, keep the compiler's call sites on the old form, and list them in the outcome for the end-of-batch seed update.

## Acceptance

- [ ] For each function: a program sample of the failure case returning `None` or `Err`, and no panic path left in its code.
- [ ] `Maps.update` has a sample; a compound assignment on a Map index is a compile error.
- [ ] The SPEC.md table lists every function with its new result type.
- [ ] `sh scripts/test.sh` passes.
