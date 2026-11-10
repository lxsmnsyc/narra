# Log: loop-unique-hoist

- Branch feat/loop-unique-hoist reset to feat/batch-5 (e13504a5). Not already fulfilled: the loops emit `_at` per write.
- Reproduced on base first. Base compiler: `sh selfhost/bootstrap.sh` on an archive of e13504a5, stage 2 C built with `cc -O2`. Note: the `narrac` that bootstrap leaves is stage 1 when stage 1 and stage 2 C match, and it carries the seed's older runtime, so I rebuilt from stage2.c.
- Base particles: 2.61G instructions in 6 of 6 runs (spread under 0.1%).
- Hoist: scan the loop (condition, step, body) for uses of each local array; only `a[expr]`, `Arrays.length(a)` and writes rooted at `a[...]` are allowed. Emitter field `hoistedArrays` (emitstate.nra, outside the brief's shared paths: one field with a default) holds the hoisted locals while the loop body is emitted.
- Not hoisted: generator bodies, locals captured by a closure, `Reference` bindings, non-mutable parameters, locals not yet declared at the loop, any bare use of the array.
- Samples: `testdata/loop_unique_generator.nra` (prints `1 2 3 2 3`), `testdata/loop_unique_hoist.nra` (hoisted and not hoisted cases; same output on base and branch).
- Measurements are in outcome.md.

## 2026-11-10: hoist ran before loops that run zero times

- Finding: `hoistArrayUniques` made the array unique before the loop, so a shared array was copied even when the loop ran zero times. The debug copy check reported it.
- Proof on the base (feat/batch-5 at 94943913, compiler built from an archive of it, stage 2 C built with `cc`): `NARRAC=<base narrac> sh scripts/test.sh --suite programs --filter copy_hoist` fails `programs/copy_hoist_empty_loop.nra/debug`: want exit code 0, got exit code 103, `copy check: copies on write that no copy(...) explains: array 3` (one per loop form: range, while, C-style).
- Fix: the choice of arrays (`chooseHoistedArrays`) is split from the emission. Range loops emit the unique under `if (<first condition>)` before the loop. While and C-style loops declare `int firstN = 1;` before the loop and emit the unique after the first condition check, under `if (firstN) { firstN = 0; ... }`. A C-style loop without a condition emits it before the loop.
- Branch: the same command passes (debug and release); `--filter loop_unique` passes.
- Instructions retired, release builds, 6 alternating base/branch pairs (base is feat/batch-5, which has the hoist): particles base 1.512G to 1.517G, branch 1.513G to 1.515G; sieve base 0.503G to 0.527G, branch 0.505G to 0.523G. The hoist gains hold.
