# Log: array-write-cold-paths

- Base: feat/batch-5 at 48823a59. Apple M2, load average 3 to 5 at the base runs (16 later, during the branch runs). Numbers are instructions retired from `/usr/bin/time -l`, median of 5 (3 for the compiler).
- Cost on base: sieve 2270M, particles 2612M, `narrac emit selfhost` 2628M and 2639M. Wall (median of 15): sieve 0.261 s, particles 0.110 s.
- Proof without code change: `variant1.py` from profile-programs applied to the base sieve C gives 710M instructions. Cost reproduced.
- Change: panic is `_Noreturn` and cold; `narra_bounds` is an inline unsigned compare plus a cold `narra_bounds_fail`; array `_unique` is an inline `rc == 1` test plus a noinline `_copy`. GNU attributes sit behind macros in runtime.h.
- Box and map splits measured on the compiler: a compiler C with all splits ran 2519M and 2515M instructions, the same C with box and map splits removed 2525M and 2513M. No gain, so the box and map `_unique` stay as they were.
- Seed regenerated twice (runtime text edit), then `bootstrap.sh` reports it current.
