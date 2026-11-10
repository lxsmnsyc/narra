# Outcome: array-write-cold-paths

## Built

- `selfhost/runtime/runtime.h`: macros `NARRA_COLD`, `NARRA_NOINLINE`, `NARRA_LIKELY`, `NARRA_UNLIKELY` (GNU attributes behind `#if defined(__GNUC__)`, empty otherwise). `narra_panic` is `_Noreturn` and cold. `narra_bounds` is an inline unsigned compare `(uint64_t)i >= (uint64_t)len`; the message formatting moved to the cold `narra_bounds_fail`.
- `selfhost/src/emitdecls.nra`: array `_unique` is an inline `rc == 1` test; the copy is `<array>_copy`, noinline.
- Generated files `selfhost/src/runtimecore.nra` regenerated; `seed/narrac.c` regenerated twice (runtime text edit). `sh selfhost/bootstrap.sh` reports the seed current.
- Box and map `_unique` are NOT split (see Gotchas).

## Proof

Apple M2, clang 21, base feat/batch-5 at 48823a59. Instructions retired from `/usr/bin/time -l`, median of 5 (3 for the compiler). Load average was 3 to 5 for the base runs and 16 to 42 for the branch runs, so wall times are not comparable across the two; instruction counts do not depend on load.

- base: sieve 2270M instructions, 0.261 s (median of 15, load ~5); particles 2612M, 0.110 s; `narrac emit selfhost` 2628M, 2639M, 2693M (three sessions).
- branch: sieve 717M instructions (-68%); particles 2623M (unchanged, as expected: that gain is loop-unique-hoist); `narrac emit selfhost` 2630M against 2693M for the base measured in the same session (-2.3%), 2559M against 2639M in an earlier session with the box and map splits still in (-3%). Wall time (median of 15, load 13): sieve 0.157 s against 0.335 s for base in the same session. Under load 33 to 42: sieve 0.339 against 0.607 s. The sieve gain is far above the spread (instruction spread under 1%).
- Without the code change the cost was reproduced on base by applying `profile-programs/experiments/variant1.py` to the base sieve C: 710M instructions.
- Out-of-range sample `testdata/panics/bounds.nra`: same output on base and branch, `panic at bounds.nra:4:15: index 5 out of range for length 2`, exit 101.
- `clang -std=c11 -pedantic -Wall` on the emitted sieve C: 138 warnings on base and on branch (same count, none new).

## Deviations

- Box and map splits left out: the brief says to keep any that do not help out of the change (measurement under Gotchas).
- Wall-clock rows for sieve and particles were taken on a busy machine; the orchestrator should record them on an idle one. `bench/BASELINE.md` untouched.

## Stubs left

- None.

## Gotchas

- Box and map splits measured on the compiler: the compiler C with all splits ran 2519M and 2515M instructions; the same C with box and map `_copy` not noinline ran 2525M and 2513M. No difference, so `<box>_unique` and `<map>_unique` keep their single function.
- particles does not change here (2612M to 2623M); its gain needs the hoist in loop-unique-hoist, which still has the per-element `rc` test and header reloads to remove.
- A text replacement on the compiler's emitted C also matches `RUNTIME_H`; the change here was made in the sources instead.
- Runtime text edit: `sh scripts/test.sh --suite generated --update`, then `sh scripts/seed-update.sh` twice.

## Proposed learnings

- Runtime helpers called in hot loops keep failure and copy paths out of line behind macros | sieve spent 90% of its samples in a non-inlined `_at` | selfhost/runtime | no

## Verification

- `sh scripts/test.sh`: PASS (8 shards and the bootstrap job, ASan and leak check included).
- `sh selfhost/bootstrap.sh`: stage 1 and stage 2 emit identical C, seed current.
- `narrac fmt selfhost/src/emitdecls.nra`: no change.

## Size

- Estimated S, actual S.
