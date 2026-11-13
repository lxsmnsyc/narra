# Outcome: simd-design

## Built

- D112 in `DECISIONS.md`, "Proposed, waiting for the maintainer". Answers:
  1. Types: twenty fixed types, ten lane types (`i8` to `u64`, `f32`, `f64`) at 128 and 256 bits. Widths do not depend on the target. No generic `Vector<T, N>`, no 512-bit types, no separate mask types.
  2. Operations: `+ - *` and unary `-` lane by lane, `/` on floats, bitwise on integers, wrapping; no scalar broadcast (`Simd.splat`). `==` gives a whole-vector `bool`; lane comparisons are `Simd.less` and the others, giving the signed integer vector of the same shape as the mask; `Simd.select`, `any`, `all`. `v[i]` reads lanes, `Simd.withLane` replaces one. `Simd.load` returns `Option`, `Simd.store` returns `bool`; FixedArray through `fromFixed`/`toFixed`. `Simd.shuffle` with constant lanes. `sum` in a fixed order, `min`, `max`, `abs`, `shiftLeft`, `shiftRight`. No conversions, `sqrt` or two-vector shuffles yet.
  3. Portability: C vector extension typedefs with `aligned(8)` (runtime heap blocks are 8-byte aligned), wrapping through unsigned twins, `memcpy` loads, no `-march`; TCC cannot build `std/simd` programs.
  4. No effects, no panics, no allocation; lane indexes use the FixedArray proof (E0994), shift counts the scalar proof (E0980).
  5. `bench/dot`: F32x8 dot product, Go and Rust with eight accumulators in the same order.
- SPEC: new subsection "Vector types" in section 4, plus the operators bullet, the module list, the `Ptr` target list and section 11, each marked not built yet.
- `selfhost/src/diagnostics.nra` header: `E1250-E1269  SIMD vector types (std/simd)`.
- Briefs: `nodes/simd/brief.md` refined (types, literals, operators, lanes, M, E1250-E1255); new `nodes/simd-library/brief.md` (functions and bench, M, E1260-E1261, depends on simd).

## Proof

- Measured on C written as the emitter will write it (`/private/tmp/claude-501/simd-design/dot.c`, `optload.c`, arm64, clang 21 `-O2`): checked scalar 0.288 s, unchecked scalar 0.198 s, F32x4 0.049 s, F32x8 0.025 s, F32x8 through an Option load 0.036 s.

## Deviations

- The build brief was split into two M nodes (simd, simd-library) instead of one, to keep each at M. The original id `simd` is the first child.
- No fenced SPEC samples: a sample of an unbuilt feature needs a `spec-skip.txt` line, outside this node's paths. The samples are in the build briefs.

## Stubs left

- None.

## Gotchas

- Vector typedefs must keep `aligned(8)`: an `Array<F32x4>` element at an 8-byte boundary is undefined behavior with the natural alignment.
- C signed vector overflow is undefined; integer `+ - *` must go through the unsigned twin.
- gcc was not run; the Linux check at the end of the batch is the first gcc build.

## Proposed learnings

- Vector types in emitted C carry `aligned(8)` | runtime heap blocks start 8 bytes past a 16-byte boundary | selfhost/src | yes (an ASan sample over `Array<F32x8>`)

## Verification

- `sh scripts/test.sh --suite sources`: PASS, 1 case. `--suite spec`: PASS, 62 cases.

## Size

- Estimated S, actual S.

## Questions for the maintainer

1. Twenty types at 128 and 256 bits, no 512 (recommended), or only the three SPEC names?
2. Lane comparisons as functions returning integer masks, with `==` a whole-vector `bool` (recommended), or `<` and `==` as mask operators?
3. `Simd.store` returning `bool` (recommended), or `Result<void, string>` so an ignored failure is a compile error?
4. No `-march` flag, SSE2 and NEON only (recommended), with wider instructions later through build targets?
