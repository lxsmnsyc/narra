---
id: const-float-exact
title: "@const evaluation reads long float literals"
status: active
priority: could
type: fix
story: batch-20
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/const-float-exact/**
shared:
  - selfhost/src/check/consteval.nra
  - testdata/const_*
  - selfhost/tests/src/**
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---
# `@const` evaluation reads long float literals

## Goal

`const t = twice(0.49999999999999994)` with an `@const` function fails: "constant t cannot be computed while compiling: it uses the float literal 0.49999999999999994, which has more digits or a larger exponent than the compiler reads exactly" (main, 2026-11-19). The same literal compiles and prints correctly at run time, and as a plain `const`. A literal should be accepted when it reads back to a finite double (correctly rounded decimal to binary64), in `selfhost/src/check/consteval.nra`.

Also from the inbox, if it falls out of the same code: `F32.from(int)` beyond 2^53 is E1052 in a `@const` body (double rounding); a direct int-to-f32 conversion would let it fold. Skip it if it needs more than a small change, and say so in the outcome.

## Acceptance

- [ ] A case with `0.49999999999999994`, `1e300` and `5e-324` in `@const` evaluation fails on the batch base and passes on the branch, and its folded values match the run-time values bit for bit.
- [ ] `sh scripts/test.sh` passes.
