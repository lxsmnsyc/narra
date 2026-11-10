# Outcome: effect-values-skip

## Built

- `selfhost/src/effects.nra`: `EffectState.storeBodies` (`Array<bool>`, one entry per `Checker.bodies`, true when a `FunctionStore` names that body), filled in `checkEffects`. `walkExpression` records into `valuesAt` only when the body being walked has a store. The only reader, `findStoredValue`, runs for store bodies, so its results are unchanged.

## Proof

- base: 46cfc1ea, `/usr/bin/time -l narrac emit selfhost`, instructions retired, 8 alternating pairs, load average 15 to 22. Mean 2567.3M, range 2542.4M to 2589.4M (spread 1.8%).
- branch: same command. Mean 2515.4M, range 2500.0M to 2544.9M. Gain 2.0% of instructions; the branch won all 8 pairs. The ranges overlap by 2.5M (0.1%) at the edges, but each pair favors the branch. Raw numbers are in `log.md`.

## Deviations

- The seed rejects a non-constant field default, so the set is an array indexed by body, not a map with a default. `lspnavigation.nra` builds an `EffectState` literal with `report: false`, so its empty `storeBodies` is never read.

## Stubs left

- None.

## Gotchas

- A new `EffectState` literal elsewhere must keep `report` false or fill `storeBodies` (one `false` per body), or `walkExpression` indexes past the end.

## Proposed learnings

- None.

## Verification

- `narrac emit selfhost` from the base and branch compilers: identical C.
- `narrac fmt selfhost/src/effects.nra`: no change.
- `sh scripts/test.sh` (NARRAC set to the branch compiler built from the seed): PASS, all shards and the bootstrap job.

## Size

- Estimated S, actual S.
