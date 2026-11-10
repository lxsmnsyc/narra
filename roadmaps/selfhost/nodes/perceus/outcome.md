# Outcome: perceus

## Built

- Step 1, borrowed for-of items: `emitArrayFor` (`selfhost/src/emitstmts.nra`) no longer retains each item or releases it at the end of the iteration. The loop already holds the array (consumed, so retained when borrowed); items are read-only, and a write to the source array copies it first because the loop's reference keeps the count above one.
- Step 2, borrowed `is` and `match` bindings: `bindPattern` (`selfhost/src/emitmatch.nra`) takes `borrowed`. `match` and `if (x is P(y))` bindings copy the value at their path with no retain and no release; guarded or-alternatives no longer release on a failed guard. The subject is held for the statement: an owned temporary already is; a local that cannot be assigned (`const`, a plain parameter, a loop item or another binding; `isStableSubject`) needs nothing; any other subject is retained once when an arm binds a counted value (`hasCountedBinding`). `emitIfIs` now skips its retain for a local that cannot be assigned. Destructuring (`const { a } = v`) still retains, because its locals outlive the statement.
- The rules are in the ownership comment at the top of `selfhost/src/emitstate.nra`, in `compiler/README.md` (reference counts) and in D55's cost line.
- Samples: `testdata/perceus_loop_items` (the body reassigns or pushes to the array it iterates, a loop over a struct field the body replaces, entries with break and continue, enum items matched while the source changes, an item returned from the loop), `testdata/perceus_match_bindings` (an arm and an `if is` body reassign the subject, a field subject whose owner is replaced, a list walk that reassigns the subject, a binding captured by a closure, guarded or-patterns, a tail call), `testdata/perceus_generators` (the same while suspended at a yield, and generators dropped mid-loop).
- Step 3, last-use moves, was built, tested and measured, then dropped (see Proof). The patch is `roadmaps/selfhost/nodes/perceus/last-use-moves.patch` (scan `findMoves`, `takeMove`, the `perceus_moves` sample).

## Proof

`bench/run.sh -n 9`, medians in seconds, load is the 1-minute average at start / end.

- base (feat/batch-4, 55fb951): interp 0.392, compiler 0.615 (load 1.91 / 2.12). Same session as the branch run below: fib 0.112, sieve 0.261, matmul 0.111, strings 0.233, interp 0.392, particles 0.188, compiler 0.604 (load 1.95 / 2.01).
- branch: fib 0.112, sieve 0.269, matmul 0.111, strings 0.243, interp 0.156, particles 0.187, compiler 0.579, peak memory 128 MB (load 1.88 / 2.17). A second branch run after the base run: fib 0.111, sieve 0.267, matmul 0.111, strings 0.235, interp 0.153, particles 0.187, compiler 0.577, 127 MB (load 1.96 / 2.27).
- fib, sieve, matmul, strings and particles compile to the same C on both sides, so their differences (sieve 0.261 to 0.269, strings 0.233 to 0.243) are noise.
- Where the time went: the emitted C of `bench/interp/interp.nra` differs from the base only by the `retain`/`release` of the `Op` item in the 150M-iteration for-of loop. interp is now 0.156 s, under its 0.237 s before D55; the acceptance target (within 10% of 0.237) is beaten, not matched.
- Compiler: 0.604 to 0.579 s wall (4%). CPU time of `narrac emit` on one source under load 4 to 6: 15 interleaved rounds, medians, base 0.600, step 1 alone 0.610 (noise), steps 1 and 2 0.580; 25 rounds, means, base 0.606, steps 1 and 2 0.576. The self-build has 9663 retain lines against about 13500 at the base, and 174058 C lines against 183740.
- Step 3: on top of steps 1 and 2 it removed 677 more retain lines from the self-build, but `narrac emit` took 0.599 s CPU against 0.576. The same code built without running the scan took 0.573, so the moves in the compiler's own code save under 1% and the scan costs about 4%. Not worth 300 lines in the ownership core.
- Soundness, with stage 1 mutants (seed-built compilers with one rule broken): no loop hold, `perceus_loop_items` and `perceus_generators` read freed memory; no subject hold in `match`/`is`, `perceus_match_bindings` reads freed memory; no subject hold in `is` alone, the same.

## Deviations

- Step 3 is not shipped (reason above). The brief's sample list for it (a moved local inside a loop, an early `return` and a `break` after a move) is in the patch, not in `testdata/`.
- The seed is not regenerated: the worker instructions for this batch leave that to the orchestrator, which regenerates after perceus.
- `sh scripts/linux-check.sh` was not run; the batch runs it once on the story branch.
- `emitMapFor` (`selfhost/src/emitmaps.nra`, not a shared path) still retains Map.keys and Map.entries items; the same borrow applies there.

## Stubs left

- None.

## Gotchas

- A borrowed binding is only as safe as the subject hold. Code that adds a new way to bind from a subject (a new pattern site) must hold the subject for the bindings' scope or pass `borrowed: false`.
- `isStableSubject` treats parameters of self tail-call functions as stable. They are reassigned only at the jump, after the arguments are consumed and everything is unwound, so a borrowed binding never outlives them.
- With moves, the subject of a `match`/`is` must not be moved while its bindings live; the dropped scan avoided it by never moving inside nested blocks. A later move pass needs the same rule.
- The compiler row is noisy under load (0.65 to 0.90 s at load 3 to 6 for the same binary); compare CPU time of interleaved runs instead.

## Proposed learnings

- Measure a compile-time analysis by building the same code with the analysis switched off: the analysis cost can cancel the gain | the moves scan cost 4% and saved under 1% | selfhost/src | no
- Hold a value for the whole statement before borrowing from it, and check with a sanitized sample that drops every other owner mid-statement | borrowed bindings read freed memory without the hold | selfhost/src | yes

## Verification

- `NARRAC=<built narrac> sh scripts/test.sh --suite programs` and `--suite bootstrap` after each step: pass.
- `sh scripts/test.sh` (full fast check, including the sanitized bootstrap job) on the final tree: PASS in 81 s. Stage 2 and stage 3 emit identical C (174058 lines).

## Size

- Estimated L, actual L. Step 3 took most of the work and was dropped after measuring.
