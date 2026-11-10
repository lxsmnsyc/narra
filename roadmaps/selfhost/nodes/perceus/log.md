# Log: perceus

All numbers are `bench/run.sh -n 9` medians in seconds unless noted. Load is the 1-minute average at the start and end of the run. Each tree was benchmarked from an archive in the scratchpad, so edits in the worktree did not leak into a run.

## Baseline (feat/batch-4, 55fb951)

- interp 0.392, compiler 0.615 (129 MB). Load 1.91 / 2.12. Matches BASELINE.md (0.398, 0.616 at 4518a91).
- Emitted C of `bench/interp/interp.nra`: the hot loop `for (const op of Arrays.values(program))` does `retain_E_..._Op(v_op_13)` at the top of the body and `release_E_..._Op(v_op_13)` at the end. `step` matches a borrowed parameter with `int` bindings, so the loop body is the call plus that pair, 150M times. Step 1's C differs from the base in exactly those two lines.

## Step 1: borrowed for-of items

- `emitArrayFor` (emitstmts.nra) no longer retains the item or registers it for release; the loop already holds the array (retained when borrowed).
- interp 0.156 (load 1.95 / 2.85).
- Round under load: interp 0.155, compiler 0.706 (load 3.78 / 3.30). Base in the same queue: interp 0.396, compiler 0.647 (load 3.94 / 3.67).

## Step 2: borrowed `is` and `match` bindings

- `bindPattern` takes `borrowed`; `is` and `match` bindings no longer retain or release. `emitMatchArms` holds the subject when it is not an owned temporary or a local that cannot be assigned and an arm binds a counted value (one retain per match instead of one per binding). `emitIfIs` skips its retain of a local that cannot be assigned. Guarded or-alternatives no longer release on a failed guard. Destructuring still retains.
- Emitted self-build: 183740 lines and about 13500 retain lines at the base; 174058 lines and 9663 retain lines after steps 1 and 2.
- Round under load: interp 0.158, compiler 0.902 (load 2.79 / 4.37). The compiler row is noise at this load; see the CPU time comparison below.

## Step 3: last-use moves (built, measured, dropped)

- Built: a scan per function body (`findMoves`) that marks the last read of a `const`/`let` local when it is a statement of the local's own block, not in a loop condition, closure, match, `&&`/`||` right side or a builtin argument other than Array.push, and the only use in its statement. The read becomes an owned temporary and the local leaves its scope frame. Samples `perceus_moves` (moved local inside a loop, early return, break, branch-only use, closure capture, short-circuit) passed under ASan and the leak check, and the sanitized bootstrap suite passed.
- Static effect on the self-build: retain lines 9663 to 8986, release lines 23128 to 20179.
- Wall clock under load: interp 0.159, compiler 0.749 (load 3.67 / 6.17).
- CPU time of `narrac emit` on the same source (user + sys, 25 interleaved rounds, mean, load 4 to 5):
  - base 0.606, steps 1-2 0.576, step 3 0.599.
  - step 3 code built without running the scan: 0.573. Steps 1-2 built the same way (`narrac build -release`): 0.577.
  - So the moves in the compiler's own code save under 1% (noise), and the scan costs about 4%. Step 3 is not worth its 300 lines in the ownership core. Dropped; the patch is `last-use-moves.patch` in this folder.

## Soundness checks (mutants)

Each mutant is a stage 1 compiler built by the seed from a source with one rule broken, then used to build the samples with `-sanitize`:

- Loop does not hold the array: `perceus_loop_items` and `perceus_generators` report heap-use-after-free.
- `match` and `is` do not hold a non-stable subject: `perceus_match_bindings` reports heap-use-after-free.
- `is` alone does not hold the subject: `perceus_match_bindings` reports heap-use-after-free.
- Step 3, a use inside a nested block or loop may move: `perceus_moves` reports heap-use-after-free in lstr_release.
- Step 3, the one-use-per-statement rule removed: no sample fails. The scan already visits an assignment target after its value and never moves a target, so the rule is defensive only.

## Final (steps 1 and 2), all programs, alternating with the base

- branch: fib 0.112, sieve 0.269, matmul 0.111, strings 0.243, interp 0.156, particles 0.187, compiler 0.579 (load 1.88 / 2.17).
- base: fib 0.112, sieve 0.261, matmul 0.111, strings 0.233, interp 0.392, particles 0.188, compiler 0.604 (load 1.95 / 2.01).
- branch again: fib 0.111, sieve 0.267, matmul 0.111, strings 0.235, interp 0.153, particles 0.187, compiler 0.577 (load 1.96 / 2.27).
- Only interp's C differs between base and branch among the six programs, so the sieve and strings differences are noise.
- `sh scripts/test.sh`: PASS in 81 s.
