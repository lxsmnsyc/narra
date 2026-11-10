# Outcome: loop-unique-hoist

## Built

- `selfhost/src/emitstmts.nra`: `emitWhile`, `emitCFor`, `emitRangeFor` and `emitRangeValueFor` scan the loop (condition, step, body) with `scanHoistExpression` / `scanHoistStatement`. For each local array that the body writes by index (`a[i] = ...`, `a[i].field = ...`, `a[i][j] = ...`, compound forms), `hoistArrayUniques` emits `*p = A_unique(*p);` once before the loop and `restoreHoistedArrays` drops the mark after it.
- The array is hoisted only when every use in the loop is `a[expr]`, `Arrays.length(a)` or a write rooted at `a[...]`; the local is not captured by any closure, not a `Reference` binding, is declared before the loop (or is a `mutable` parameter); and the loop is not in a `function*` body.
- `selfhost/src/emitexprs.nra`: `indexPointer` emits `narra_bounds(i, (*a)->len, where); &(*a)->items[i]` for a hoisted array, with no `rc` test.
- `selfhost/src/emitstate.nra`: field `hoistedArrays: Array<int> = []` on `Emitter`.
- Samples: `testdata/loop_unique_generator.nra` (prints `1 2 3 2 3`), `testdata/loop_unique_hoist.nra` (hoisted loops, and loops that copy into a binding, pass the array on, push it into a container, capture it in a closure, or replace it; same output as on base).

## Proof

Apple M2, base feat/batch-5 at e13504a5 (compiler built from an archive of that commit, stage 2 C built with `cc`), branch compiler from this tree. Alternating base/branch pairs, 6 each, `bench/<program>/<program>.nra` built with `build -release`, instructions from `/usr/bin/time -l`. Load average 10 to 21 (other workers running), so wall times are noisy; instruction counts do not depend on load.

- particles: base 2.612G to 2.615G instructions in 6 of 6 runs; branch 1.512G to 1.513G (-42%); branch won 6 of 6 pairs by instructions. Wall time at the quiet end of the run: 0.11 s base, 0.06 to 0.07 s branch.
- sieve: base 0.689G to 0.754G (median 0.695G); branch 0.497G to 0.526G (median 0.518G) (-25%); 6 of 6 pairs. Wall time 0.15 to 0.17 s on both sides (memory bound, as predicted).
- `narrac emit selfhost` (same source, 3 runs per side): base 2.493G, 2.448G, 2.443G; branch 2.508G, 2.499G, 2.499G (+0.6% to +2.3%). The scan costs a little in the compiler and the compiler's own loops gain nothing measurable. Reported as asked; no gain was required.
- Samples: `loop_unique_generator` prints `1 2 3 2 3` with the branch compiler; `loop_unique_hoist` output is identical on base and branch.

## Deviations

- `selfhost/src/emitstate.nra` is outside the brief's shared paths: the hoisted set has to live on `Emitter`. One field with a default; the orchestrator should expect a possible textual conflict only if another node edits the same lines.
- The scan lives in `emitstmts.nra` (no new file).

## Stubs left

- None. `for (x of Arrays.values(a))` loops and the array `for` forms are not hoisted (the brief names while, C-style and range loops).

## Gotchas

- After a bootstrap where stage 1 and stage 2 C are equal, the `narrac` left in the output directory is stage 1, which carries the seed's older runtime. To measure a base compiler, build `stage2.c` with `cc`.
- A closure that is passed straight to a call inside the loop and captures the array counts as a capture, so the array is not hoisted.
- The unique is made before the loop even if the loop runs zero times; the copy is not observable.

## Proposed learnings

- Measure a base compiler from its own stage 2 C, not from the `narrac` bootstrap leaves | stage 1 carried the older runtime and made the base look 3x slower on sieve | bench, selfhost | no

## Verification

- `narrac fmt` on the three edited `.nra` files: no change.
- `sh selfhost/bootstrap.sh`: fixed point reached (stage 2 and stage 3 emit identical C); seed not regenerated, as instructed.
- `sh scripts/test.sh`: PASS (8 shards of 733 and the bootstrap job; sanitizers and leak check included). The first run failed only on the new sample's missing `.out` and its format; both fixed.

## Size

- Estimated M, actual M.
