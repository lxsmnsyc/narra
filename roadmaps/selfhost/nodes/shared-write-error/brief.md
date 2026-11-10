---
id: shared-write-error
title: "Shared writes become errors"
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
  - roadmaps/selfhost/nodes/shared-write-error/**
shared:
  - selfhost/src/**
  - selfhost/tests/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: shared-write-sweep
    kind: hard
    why: the compiler must be clean first
---

# Shared writes become errors

## Goal

D94 step 5 (debate settled structure items 1 and 4, P18, O18):

- The shared-write warning (E0978) becomes an error. Keep the fix.
- Turn the debug copy check on for every program sample: `isCopySample` true for every program, or `COPY_CHECK_ENVIRONMENT` in `Shell.leakCheckEnvironment()` (see `nodes/copy-builtin/outcome.md`). Any report that remains is a hidden copy to fix or explain in the outcome.
- Error samples for each D2 form with all-scalar structs, which the runtime check cannot see (copy-builtin outcome, Gotchas: borrowed values and inline all-scalar structs).

## Acceptance

- [ ] Every shared-write warning sample from shared-write-check is an error sample.
- [ ] Every program sample runs with the copy check on and reports nothing, or carries a `.copies` file whose line the outcome explains.
- [ ] One error sample per D2 form with an all-scalar struct.
- [ ] `sh scripts/test.sh` passes.
