# Log: loop-unique-hoist

- Branch feat/loop-unique-hoist reset to feat/batch-5 (e13504a5). Not already fulfilled: the loops emit `_at` per write.
- Reproduced on base first. Base compiler: `sh selfhost/bootstrap.sh` on an archive of e13504a5, stage 2 C built with `cc -O2`. Note: the `narrac` that bootstrap leaves is stage 1 when stage 1 and stage 2 C match, and it carries the seed's older runtime, so I rebuilt from stage2.c.
- Base particles: 2.61G instructions in 6 of 6 runs (spread under 0.1%).
- Hoist: scan the loop (condition, step, body) for uses of each local array; only `a[expr]`, `Arrays.length(a)` and writes rooted at `a[...]` are allowed. Emitter field `hoistedArrays` (emitstate.nra, outside the brief's shared paths: one field with a default) holds the hoisted locals while the loop body is emitted.
- Not hoisted: generator bodies, locals captured by a closure, `Reference` bindings, non-mutable parameters, locals not yet declared at the loop, any bare use of the array.
- Samples: `testdata/loop_unique_generator.nra` (prints `1 2 3 2 3`), `testdata/loop_unique_hoist.nra` (hoisted and not hoisted cases; same output on base and branch).
- Measurements are in outcome.md.
