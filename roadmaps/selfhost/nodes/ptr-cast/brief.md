---
id: ptr-cast
title: "Ptr.cast and Ptr.element in std/ptr"
status: active
priority: should
type: feature
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
  - roadmaps/selfhost/nodes/ptr-cast/**
  - selfhost/src/stdptr.nra
  - selfhost/src/emitpointers.nra
shared:
  - selfhost/src/unsafecode.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/builtins.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/fmt.nra
  - selfhost/src/trivia.nra
  - testdata/**
  - SPEC.md
  - docs/guide/**
reserves: []
provides: []
depends_on: []
---

# Ptr.cast and Ptr.element in std/ptr

## Goal

Inbox, from the batch-8 review: `std/ptr` has no cast and no element-address function, so a `Ptr<u8>` into a FixedArray buffer can only come back from a C call (the `unsafe_fixed_arrays` sample uses the address `memset` returns). Add, callable only inside `unsafe` like the other `Ptr` functions:

- `Ptr.cast<U>(pointer: Ptr<T>): Ptr<U>`;
- the address of a FixedArray element (`Ptr.element(array, index)` or the name that fits D95 and the batch-8 code), with the index checked against the length like other element reads.

Then rewrite `unsafe_fixed_arrays` to take the address directly. Also from the same review: in `from /* x */ "c"`, `narrac fmt` moves the comment before `from`; keep it in place.

## Acceptance

- Samples for both functions, clean under AddressSanitizer; error cases for each outside `unsafe` and for an out-of-range constant index.
- A fmt case keeps the comment after `from`.
- SPEC and the guide list both functions.
- `sh scripts/test.sh` passes.
