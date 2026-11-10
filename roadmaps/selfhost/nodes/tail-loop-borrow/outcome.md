# Outcome: tail-loop-borrow

## Built

- `selfhost/src/emittail.nra`: `openTailLoop` finds, with `findPassedThrough`, the parameters that every self tail call hands back unchanged. A counted one of those gets no entry retain, no `_held` variable and no release at return; it stays borrowed from the caller. `emitTailJump` skips a borrowed parameter (`isBorrowedParameter`: counted and not registered in any scope). Parameters that a jump replaces, mutable parameters and uncounted parameters are emitted as before.
- `findTailCalls` returns the tail call expressions, scanned once; `collectCallIds` turns them into the ids stored in `Emitter.tailCalls`.
- Sample `testdata/tail_loop_borrow.nra` (+ `.out`): pass-through string and array parameters next to replaced and swapped ones, returning a borrowed parameter, storing it, and a void tail loop.

## Proof

- base: 24aec0d7 (feat/batch-5), `/usr/bin/time -l narrac emit selfhost`, instructions retired, 5 interleaved runs, load average 9 to 13: 2.621e9, 2.591e9, 2.566e9, 2.557e9, 2.562e9 (median 2.567e9).
- branch: same command: 2.175e9, 2.066e9, 2.059e9, 2.054e9, 2.065e9 (median 2.064e9). Gain 19.6%, far above the spread (under 2% after the first run). Larger than the profile's 8% estimate; the stage 2 C is also 306 lines shorter.

## Deviations

- `selfhost/src/emitconsts.nra` (outside the owned paths) changed in a few lines at the call site: `findTailCalls` now returns the call expressions, `collectCallIds` makes the ids, and `openTailLoop` takes the list so the body is scanned once. One comment there was reworded to match.
- Large parameters that are not counted still use `_held`; only counted parameters are borrowed, so the jump can tell a borrowed parameter from an owned one by scope registration alone.

## Stubs left

- None.

## Gotchas

- A counted parameter that is passed through is no longer in the function's scope locals; anything that assumes every parameter of a tail-loop function is registered must not.
- The seed was not regenerated; the gain shows in the bootstrapped compiler, and in the seed after the batch's seed update.

## Proposed learnings

- A tail-recursive helper that passes a counted parameter back unchanged borrows it. | Entry retains of whole structs were 9% of `narrac emit selfhost`. | selfhost/src | no

## Verification

- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C (164026 lines).
- `NARRAC=<built> sh scripts/test.sh --filter tail_loop` and `--filter params_pointer`: pass (debug and release, sanitizers and leak check).
- Full `sh scripts/test.sh`: pass (see the final report).

## Size

- Estimated M, actual S.
