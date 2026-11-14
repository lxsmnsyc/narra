# Outcome: simd-library

## Built

- The rest of `std/simd` (D112 item 2), registered in `SIMD_FUNCS` (`selfhost/src/stdsimd.nra`), checked there and emitted in `selfhost/src/emitsimd.nra`. Every check kind starts with `checkSimd`, so a call type argument becomes the expected type.
  - `Simd.equal`, `notEqual`, `less`, `lessOrEqual`, `greater`, `greaterOrEqual` (`checkSimdCompare`): two vectors of one type, result the mask type. C: `((narra_i32x4)(a < b))`.
  - `Simd.select(mask, ifTrue, ifFalse)` (`checkSimdSelect`): E1261 when the mask is not the mask type of `ifTrue`. C: bit blend on the mask's unsigned twin, `(V)(((U)t & (U)m) | ((U)f & ~(U)m))`.
  - `Simd.any`, `Simd.all` (`checkSimdTest`): any integer vector, E1261 on float vectors. C: lane tests `m[i] != 0` joined by `||` or `&&`.
  - `Simd.load(items, start)` (`checkSimdLoad`): `V` from a type argument (`V`) or the expected type (`Option<V>`), E1254 (a second row with a load example) without either. Returns `Option<V>`. C: `start >= 0 && start <= len - count`, then `memcpy`.
  - `Simd.store(mutable items, start, v)` (`checkSimdStore`, `mutables: [true, false, false]`): returns `bool`. E0312 without `mutable`; E1261 when the array is not `Array<L>` of the vector's lane. C: the same range test; inside it the array is made unique (`_unique`), then `memcpy`.
  - `Simd.shuffle(v, [ ... ])` (`checkSimdShuffle`): E1260 for a list that is not an array literal, a list of the wrong length, a lane that is not a constant (literals and top-level constants count, through `findTypedConstant`), or a lane out of range. C: a compound literal of lane reads.
  - `Simd.sum(v)` (`checkSimdSum`): the lane type. C: lanes read from the unsigned twin, the high half added to the low half until one is left, each step cast to the twin's lane type, then cast to the lane type. Floats: the same order in `float`/`double`.
  - `Simd.min`, `Simd.max` (`checkSimdMinMax`): a blend on `a < b` (min) or `b < a` (max).
  - `Simd.abs` (`checkSimdAbs`): E1261 on unsigned vectors. Floats clear the sign bit (`(U)v & 0x7fff...`), so `-0.0` gives `0.0`. Integers blend `v` and the wrapping `-(U)v` on `v < 0`, so the lowest integer stays itself.
  - `Simd.shiftLeft`, `Simd.shiftRight` (`checkSimdShift`): E1261 on float vectors. The `int` count: a constant is read as an `int` and must be 0 to width - 1; any other count goes through `checkShiftCount` (`operands.nra`) with the signed lane type of the same width (E0980). C: `v >> (L)count`; a left shift runs on the unsigned twin.
- `findMaskType(vector)` in `selfhost/src/simdtypes.nra`: the signed integer vector of the same lane width and count. The checks intern the mask and twin types they emit, so `emitVectorTypedefs` writes their typedefs.
- `selfhost/src/diagnostics.nra`: rows for E1254 (load), E1260 (4 rows), E1261 (4 rows).
- Tests: `testdata/simd/library.nra`/`.out` (every function on float and integer types; load at the last full position and one past it; store out of range leaves the array unchanged; select with a mask from `less`; shuffle reversing lanes; F32x8 sum equal to the fixed-order scalar sum; abs of the lowest I32x4 lane; shifts on signed and unsigned lanes; a dot product with a scalar tail). Errors: `testdata/errors/simdlib_shuffle_lanes` (E1260: a variable lane, lane 4 on F32x4, three lanes for four, a non-literal list), `simdlib_vector_type` (E1261: shiftLeft on F32x4, abs on U32x4, an I64x2 mask for F32x4, any on F32x4, store of F32x4 into Array<f64>; E0312), `simdlib_shift_count` (E0980: an unproven name, a name proven only from above on U8x16, constants 32 and 264), `simdlib_load_type` (E1254, and an `Option<I32x4>` load from `Array<f32>`).
- `bench/dot/dot.nra`, `dot.go`, `dot.rs`, added to `bench/run.sh` (before `compiler`) and `bench/README.md`. Two `Array<f32>` of 1,048,576 elements, 200 passes, an `F32x8` accumulator through `Simd.load`, `Simd.sum` at the end; Go and Rust keep eight scalar accumulators summed in the order of `Simd.sum`. The values make every sum exact, and all three print `800` (the dot product times 32, truncated).
- SPEC section 4 "Vector types": the "not built yet" note is gone (it was the last D112 one). D112's status line says built by nodes simd and simd-library. `compiler/README.md`: the vector bullet names the C form of each function.

## Proof

- base: feat/batch-13 0a6fc652, release compiler from `sh selfhost/bootstrap.sh`; `narrac run testdata/simd/library.nra` exits 1 with `library.nra:20:35: module "std/simd" has no function "load"`.
- branch: `NARRAC=<branch> sh scripts/test.sh --suite programs --filter simd`: PASS, 10 cases (library matches its `.out` in debug with ASan, leak check and copy check, and in release).

## Deviations

- `DECISIONS.md` (not in the brief's paths): the D112 status line says built, and the Built note names simd-library's functions and the float `abs` rule. The brief's SPEC step asks for the status line.
- `Simd.abs` on floats clears the sign bit instead of the brief's blend of `v` and `-v`: the blend gives `-0.0` for `-0.0` and keeps a NaN's sign. Recorded in D112's Built note.
- `bench/dot` prints the dot product times 32 as an integer, from values whose sums are exact, so the three languages agree whatever the float formatting and FMA. The F32x8 accumulator and the fixed summing order are as the brief says.
- SPEC: section 4's note was the only "not built yet" text left for D112; the operators bullet, the module list, the `Ptr` target list and section 11 had already lost theirs in node simd.
- `Simd.any` and `Simd.all` take every integer vector, not only mask types (D112: "whether any or every lane is not 0").
- `Simd.store` of a vector into an array of another lane type is E1261 (one of this node's codes), not a type mismatch.

## Stubs left

- None.

## Gotchas

- `findTypedConstant(count, laneType)` wraps the constant to the lane type (264 on u8 lanes reads as 8). The vector shift check reads constants as `int` and passes names to `checkShiftCount` with the signed lane; a later use of that prover with a narrower type than the count's needs the same care.
- C vector shifts and compound literals need the scalar cast to the lane type; clang rejects a scalar that would be truncated when it is splatted.
- `a is Some(x) && b is Some(y)` cannot bind; load loops nest two `if`s.
- Code that never uses `std/simd` gets no new C: `narrac emit selfhost` from the base and the branch compiler is byte-identical.

## Proposed learnings

- A std check that proves a scalar operand reads its constant in the operand's own type | the prover wraps constants to the type it is given | selfhost/src | yes (simdlib_shift_count's 264 case)

## Verification

- Emit comparison: `narrac emit selfhost` of the branch source from the feat/batch-13 (0a6fc652) compiler and from the branch compiler: byte-identical (`cmp`, 241309 lines).
- Suites with the branch compiler: programs (`--filter simd`) 10, errors (`--filter simdlib`) 4, codes 1, sources 1, spec 65, all PASS.
- clang `-std=c11 -Wall -Wextra` on the C of `library.nra`: no warnings besides the unused ones from the runtime text.
- `bench/run.sh -n 5 dot` (load average about 15, other workers running): narra 0.069 s, go 0.316 s, rust 0.132 s, check ok (800).
- Scalar Narra version of the same loop (eight `f32` accumulators, bounds-checked `Array` reads; scratch only, not in `run.sh`), medians of 7 runs in three rounds on arm64: F32x8 0.037 to 0.048 s, scalar 0.126 to 0.144 s, setup alone 0.011 to 0.013 s. The loop is 3.7 to 4.4 times faster, the whole program 3.0 to 3.4 times. Target (3 times on arm64): met. Re-measure on a quiet machine for BASELINE.md.
- After the rebase on feat/batch-13 9284717b: `sh scripts/test.sh`: `PASS, 9313 cases in 411s`. `sh scripts/linux-check.sh` (colima, arm64, gcc): `test.sh: PASS, 9313 cases in 239s`, `linux-check: PASS in 664s`. x86-64 gcc was not run.

## Size

- Estimated M, actual M (about 900 lines with tests and the bench).
