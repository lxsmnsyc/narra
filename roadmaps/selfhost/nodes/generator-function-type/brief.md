---
id: generator-function-type
title: "A function* can be assigned to an annotated function type"
status: todo
priority: should
type: fix
story: batch-2
early: false
outline: false
milestone: m12-spec-gaps
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
added: 2026-11-09
spent:
owns:
  - testdata/generator_function_type*
  - roadmaps/selfhost/nodes/generator-function-type/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes go into the Narra compiler only, after the bootstrap fixed point
---

# A function* can be assigned to an annotated function type

## Goal

From the PR #89 review (2026-11-05): `const g: (x: int) => Generator<int> = count`, with `count` a `function*`, fails with "expected (int) => Generator<int>, found (int) => Generator<int>". The unannotated `const f = count` works. Likely a type identity mismatch between the generator's function type and the annotated one (effects or a generator flag). The report predates the Go removal and batch-1; check that it still happens before fixing.

## Acceptance

- [ ] Proof on the base: a sample with the annotated binding (top-level and local, and as an argument to a parameter of that type) fails, or "Not reproduced" with the evidence.
- [ ] Sample `testdata/generator_function_type*.nra` with `.out` passes, clean under AddressSanitizer.
- [ ] A real mismatch (wrong element type) still errors, with both types printed differently.
- [ ] `sh scripts/test.sh` passes.
