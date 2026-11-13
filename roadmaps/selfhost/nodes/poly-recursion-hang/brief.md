---
id: poly-recursion-hang
title: "Polymorphic recursion in generic functions makes the compiler run forever"
status: active
priority: should
type: fix
story: batch-12
early: false
outline: false
milestone: 
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - testdata/errors/poly_recursion_*
  - roadmaps/selfhost/nodes/poly-recursion-hang/**
shared:
  - selfhost/src/instances.nra
  - selfhost/src/generics.nra
  - selfhost/src/diagnostics.nra
reserves: []
provides: []
depends_on: []
---
# Polymorphic recursion in generic functions makes the compiler run forever

## Goal

A generic function `grow<T>(v: T)` that calls `grow(Grow { value: v })`, or a `Show.show` on `Grow<T>` that shows `Grow { value: value }`, makes the compiler run for over 60 s (probe p5 in the generic-implement review). The instance limit does not catch it. Make it a compile error with a code, as poly-recursion did for generic types.

## Context

- From the inbox, 2026-11-12. Happens on the batch-11 base too.
- Proof first: a test that times out (or a capped run) on the base, and reports the error on the branch.

## Acceptance

- [ ] Error tests for the function case and the trait implementation case, each finishing in under a second.
