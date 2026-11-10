---
id: const-params-pointer
title: "Pass large constant parameters by pointer"
status: done
priority: should
type: feature
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - testdata/params_pointer_*
  - roadmaps/selfhost/nodes/const-params-pointer/**
shared:
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emittail.nra
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: option-inline
    kind: hard
    why: inline enums grow the values passed by value, and both change function signatures in emitdecls.nra
  - on: profile-compiler
    kind: hard
    why: the profile says whether copying large parameters costs enough to do this; if not, the node is dropped
---

# Pass large constant parameters by pointer

## Goal

Build D50 ("large constant parameters are passed by pointer", SPEC section 11), which is not built: every parameter is passed by value today (`emitFunctionCall`, `selfhost/src/emitexprs.nra:774-822`). Pass a parameter by `const T*` when its C type is larger than a size limit (pick it from the profile, for example 16 bytes), in direct calls, function values, closures and tail calls.

## Context

- Parameters are already borrowed for reference counting, so a pointer changes no counts.
- `mutable` parameters already pass an lvalue pointer; reuse that path where it fits.
- Function values (`callFunctionValue`, `emitclosures.nra:540`) need one convention per function type, so the limit depends only on the type, never on the call site.
- Tail calls (`emittail.nra:279`) assign parameters in a loop; a pointer parameter must not alias an argument that is reassigned in the same step.
- Drop this node with a reason in its outcome if profile-compiler finds the copies are under noise.

## Acceptance

- [ ] `bench/run.sh` before and after, including the compiler row.
- [ ] Samples `testdata/params_pointer_*`: a large struct through a direct call, a function value, a closure and a tail call that swaps two large arguments.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass; D50's "not built" note removed; seed regenerated.

## Sizing

M: one calling convention change across call and declaration sites; blast radius high but mechanical.
