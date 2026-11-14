---
id: simd-library
title: "SIMD vector types: std/simd functions and bench/dot"
status: active
priority: must
type: feature
story: batch-13
early: false
outline: false
milestone: p1-syntax
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
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/simd-library/**
  - testdata/simd/library*
  - testdata/errors/simdlib*
  - bench/dot/**
shared:
  - selfhost/src/stdsimd.nra
  - selfhost/src/simdtypes.nra
  - selfhost/src/emitsimd.nra
  - selfhost/src/diagnostics.nra
  - bench/run.sh
  - bench/README.md
  - SPEC.md
  - compiler/README.md
reserves: [diag E1260-E1269]
provides: []
depends_on:
  - on: simd-design
    kind: hard
    why: D112 fixes the function list, codes and SPEC text this node builds
  - on: simd
    kind: hard
    why: the functions take and return the vector types that node simd adds
---

# SIMD vector types: std/simd functions and bench/dot

## Goal

Build the rest of `std/simd` from D112 item 2: comparisons giving masks, `select`, `any`, `all`, `load`, `store`, `shuffle`, `sum`, `min`, `max`, `abs`, `shiftLeft` and `shiftRight`. Add the benchmark program `bench/dot` from item 5, which shows the gain over scalar code.

## Context

- Source: D112 in `DECISIONS.md`, SPEC section 4 "Vector types", and the outcome of node simd. They override this brief where they differ.
- Proof first: `testdata/simd/library.nra` fails to compile on the base with node simd landed (`Simd.load` is unknown).
- Every function is pure: no effect, no allocation, no panic.
- C forms (measured in the spike, `/private/tmp/claude-501/simd-design/dot.c`, `optload.c`):
  - comparisons: the C operator on vectors (`a < b`) gives the mask type directly.
  - `select`: bit blend through the unsigned twin, `(t & m) | (f & ~m)`; works for any mask value.
  - `load`/`store`: bounds test `start >= 0 && start <= length - count` (no overflow from `start + count`), then `memcpy`.
  - `shuffle`: build the result from lane reads (`(narra_f32x4){v[3], v[2], v[1], v[0]}`); clang and gcc turn it into one shuffle. Do not use `__builtin_shufflevector` (older gcc lacks it).
  - `sum`: add the high half to the low half until one lane is left; write it out per type so the order is fixed.
  - `min`/`max`: mask from `<`, then the blend. `abs`: blend of `v` and `-v` (wrapping for integers).
  - shifts: `v << count` with a scalar count in C; the count proof reuses `isProvenCount` in `operands.nra` with the lane type.
- Mask type: the signed integer vector with the lane width and count of the compared type (`simdtypes.nra` from node simd).

## Steps

- [ ] Comparisons, `select`, `any`, `all`; E1261 for a mask of the wrong type.
- [ ] `load` (type from expected type or type argument, E1254 as in node simd), `store`.
- [ ] `shuffle` with a literal of constants; E1260 for a lane that is not a constant or is out of range, or a list of the wrong length.
- [ ] `sum`, `min`, `max`, `abs` (E1261 on unsigned), `shiftLeft`, `shiftRight` (E1261 on float vectors, E0980 for an unproven count).
- [ ] `bench/dot/dot.nra`, `dot.go`, `dot.rs`, added to `run.sh` and `bench/README.md`. Two `Array<f32>` of 1,048,576 elements, 200 passes, an `F32x8` accumulator through `Simd.load`, `Simd.sum` at the end. Go and Rust keep eight scalar accumulators summed in the same order, so all three print the same number.
- [ ] SPEC: drop the "not built yet" notes of D112 (section 4 "Vector types", the operators bullet, the module list, the `Ptr` target list, section 11). D112's status line says built.

## Acceptance

- [ ] `testdata/simd/library.nra` with checked output, clean under AddressSanitizer, the leak check and the copy check: each function on at least one float and one integer type, `load` at the last full position and one past it (`None`), `store` out of range (`false`, array unchanged), `select` with a mask from `less`, `shuffle` reversing lanes, `sum` of `F32x8` equal to the fixed-order scalar sum, `abs` of the lowest `I32x4` lane, shifts on signed and unsigned lanes.
- [ ] Error tests: E1260 (a variable lane, lane 4 on `F32x4`, three lanes for four), E1261 (`Simd.shiftLeft` on `F32x4`, `Simd.abs` on `U32x4`, `Simd.select` with an `I64x2` mask for `F32x4`), E0980 (an unproven shift count), E1254 (`Simd.load(items, 0)` with no expected type).
- [ ] `bench/run.sh dot` prints the same answer for Narra, Go and Rust. The outcome records the Narra time beside a scalar Narra version of the same loop (written for the measurement, not committed to `run.sh`); the target is at least 3 times faster on arm64.
- [ ] `sources` and `codes` suites pass.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: about 13 functions with type checks and C forms on top of node simd's types, plus one bench program in three languages. About 700 lines.

## Notes for the worker

- Codes E1260 and E1261 are yours; E1262-E1269 are spare.
- Conversions between lane types, `sqrt` and two-vector shuffles are out of scope (D112 item 2).
- The Rust version must not use nightly `std::simd`; plain arrays of eight accumulators.
