---
id: simd
title: "SIMD vector types: types, literals, operators, lanes"
status: todo
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
  - roadmaps/selfhost/nodes/simd/**
  - selfhost/src/stdsimd.nra
  - selfhost/src/simdtypes.nra
  - selfhost/src/emitsimd.nra
  - testdata/simd/**
  - testdata/errors/simd*
shared:
  - selfhost/src/types.nra
  - selfhost/src/cnames.nra
  - selfhost/src/check.nra
  - selfhost/src/exprs.nra
  - selfhost/src/literals.nra
  - selfhost/src/fixedarrays.nra
  - selfhost/src/traits.nra
  - selfhost/src/generics.nra
  - selfhost/src/unsafecode.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/stdlib.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1250-E1259]
provides: []
depends_on:
  - on: simd-design
    kind: hard
    why: D112 fixes the types, operators, codes and SPEC text this node builds
---

# SIMD vector types: types, literals, operators, lanes

## Goal

Build items 1 to 4 of D112 except the library functions: the twenty vector types of `std/simd`, array literals with an expected vector type, lane-wise operators, lane reads with proven indexes, `==`, hashing and `Show`, and the four building functions `Simd.splat`, `Simd.fromFixed`, `Simd.toFixed` and `Simd.withLane`. They compile to C vector extensions. Node simd-library adds the rest of `std/simd` and the benchmark.

## Context

- Source: D112 in `DECISIONS.md` and SPEC section 4 "Vector types" and section 11. They override this brief where they differ.
- Proof first: `testdata/simd/operators.nra` fails to compile on the batch base (`std/simd` is an unknown module).
- No new syntax. The parser, `ast.nra`, `astdump.nra` and `fmt.nra` need no change; type names go through the existing type resolution.
- Pattern to copy: `std/ptr` (`stdptr.nra`, `unsafecode.nra`): a builtin module whose type name is visible when the module is imported.
- Type: one variant `TVector(lane: Type, count: int)` in `types.nra`. Keep its logic in `simdtypes.nra` (the table of the twenty types, names, mask type, lane width) and add only dispatch rows to the shared switches (`TPtr` has about 35 sites; follow it).
- Index proof: reuse `checkFixedIndex` in `fixedarrays.nra`; make `findFixedLength` (or a sibling) return the lane count of a vector.
- C: emit a typedef per used type, `typedef float narra_f32x4 __attribute__((vector_size(16), aligned(8)));`, and its unsigned twin for wrapping. Alignment 8 is required: heap blocks start at 8-byte boundaries (`narra_block_new`). Integer `+ - *` and unary `-` cast to the unsigned twin and back. Lane reads are `v[i]`. `==` compares lanes and reduces (all lanes equal). Do not edit `selfhost/runtime/` (that needs two seed updates); write the typedefs from the emitter.

## Steps

- [ ] `TVector` type, the twenty names visible after `import * as Simd from "std/simd"`, type strings, C names, layout (size 16 or 32, align 8), `Ptr` target allowed.
- [ ] Array literal with an expected vector type; E1253 on a wrong lane count.
- [ ] Operators: `+ - *` all types, unary `-` signed and float (E0296 on unsigned), `/` float, `& | ^ ~` integer, compound forms. E1250 for `< <= > >= % << >>` and integer `/` (message names the `Simd` function); E1251 for a vector with a scalar operand (hint `Simd.splat(...)`).
- [ ] `v[i]` lane reads with the `FixedArray` proof (E0994); E1252 on `v[i] = x` and compound lane writes.
- [ ] `==`, `!=`, hash for integer vectors (float vectors fall under the float key rule), `Show` as a `FixedArray`.
- [ ] `Simd.splat` (E1254 without an expected type or type argument), `Simd.fromFixed` (E1255 when the lane type and length have no vector type), `Simd.toFixed`, `Simd.withLane` (lane proven like `v[i]`).
- [ ] Not a contents type: `mutable` parameter and `copy` errors come from the existing rules; add error tests that show them.
- [ ] `compiler/README.md`: the feature list; SPEC: drop "not built yet" for what this node builds (leave it on the simd-library functions).

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check:
  - `testdata/simd/operators.nra`: every operator on `F32x4`, `F64x4`, `I32x8`, `U8x16`; wrapping (`I32x4` lanes at `i32` max plus 1 give min; `U8x16` 255 + 1 gives 0).
  - `testdata/simd/lanes.nra`: literals, `v[2]`, a proven index in `for (let i = 0; i < 4; i++)`, `withLane`, `splat`, `fromFixed`/`toFixed` round trip, `==`, a template of a vector.
  - `testdata/simd/containers.nra`: `Array<F32x8>` built with `Arrays.push` and summed, a struct with a vector field, a `FixedArray<I64x2, 3>`, a generic `function first<T>(items: Array<T>): Option<T>` called on vectors, a `Map<I32x4, string>`.
- [ ] Error tests in `testdata/errors/`: E1250 (`a < b`, `a % b` on `I32x4`, `a / b` on `I32x4`, `a << b`), E1251 (`v * 2.0`), E1252 (`v[0] = 1.0`), E1253 (three lanes for `F32x4`), E1254 (`const x = Simd.splat(1.0)`), E1255 (`Simd.fromFixed` of `FixedArray<f32, 3>`), E0994 (`v[4]` on `F32x4`, an unproven name), E0296 (`-u` on `U32x4`), and the float map key error on `Map<F32x4, int>`.
- [ ] The emitted C of `operators.nra` compiles with `-std=c11 -Wall` without vector warnings under clang; the Linux check (gcc) passes at the end of the batch.
- [ ] `narrac query` and the language service show vector type names in hovers and signatures (a case in the `query` or `lsp` suite).
- [ ] `sources` and `codes` suites pass.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: one type variant across the existing switches plus four std functions; no syntax. About 900 lines with tests.

## Notes for the worker

- Codes E1250-E1255 are yours; E1256-E1259 are spare. simd-library uses E1260-E1269.
- Do not use the feature in `selfhost/src` (seed rule).
- `types.nra` is a hot file: keep the variant and the switch rows small; logic goes in `simdtypes.nra`.
- Float lanes compare with C `==` per lane, so a vector holding NaN is not equal to itself, as a float is not.
