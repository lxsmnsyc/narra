# Log: simd-library

## Proof first

- Base: feat/batch-13 0a6fc652, compiler built with `sh selfhost/bootstrap.sh /private/tmp/claude-501/simd-library/base` (release).
- `base/narrac run testdata/simd/library.nra` exits 1: `library.nra:20:35: module "std/simd" has no function "load"`, then the other missing functions.
- Branch: `--suite programs --filter simd` PASS, 10 cases (debug with ASan, leak check and copy check, and release).

## Notes

- `a is Some(x) && b is Some(y)` cannot bind names; the sample and the bench nest two `if`s.
- A store into an array that another name still holds is rejected by the shared-write check ("may change target, which shared still uses"), so the unique copy in `Simd.store` only runs for arrays the checker cannot see as shared. The sample does not reach it.
- Shift counts: `isProvenCount` wraps a constant to the operand type, so `Simd.shiftLeft(u8s, 264)` would pass as 8 if sent through `checkShiftCount` with the lane type. Constants are read as `int` here; names go through `checkShiftCount` with the signed lane, since an unsigned lane would accept a negative `int` count.
- bench/dot values are multiples of 1/8 and 1/4 whose sums cancel every 112 elements, so every sum is exact: the answer (800 = 25 * 32) does not depend on FMA or the order of additions. Go still converts each product (`float32(a*b)`) so it is not fused.

## Bench (load average 12 to 20, other workers running)

- `bench/run.sh -n 5 dot`: narra 0.069, go 0.316, rust 0.132, check ok (800).
- Medians of 7 runs, three rounds: setup only (passes = 0) 0.011 to 0.013; F32x8 0.037 / 0.048 / 0.045; scalar Narra with eight accumulators 0.126 / 0.144 / 0.143. The loop alone: 0.026-0.035 against 0.115-0.131, 3.7 to 4.4 times faster; the whole program 3.0 to 3.4 times.
