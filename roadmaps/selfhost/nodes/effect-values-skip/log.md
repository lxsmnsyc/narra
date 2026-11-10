# Log

- Base: 46cfc1ea (feat/batch-5 at start). Built base and branch compilers with `sh selfhost/bootstrap.sh`.
- Measured `narrac emit selfhost` with `/usr/bin/time -l`, "instructions retired", 8 alternating base/branch pairs. Load average 15 to 22 (shared machine).
- Base (instructions): 2587094616, 2564875002, 2574564055, 2566593471, 2550658928, 2589405667, 2562940641, 2542427483. Mean 2567.3M, range 2542.4M to 2589.4M.
- Branch: 2499998540, 2532031424, 2517930883, 2544856614, 2506880576, 2518791140, 2502679829, 2499963571. Mean 2515.4M, range 2500.0M to 2544.9M.
- Branch wins 8 of 8 pairs; mean gain 2.0%.
- The seed rejects a non-constant field default (`Map {}`), so `storeBodies` is `Array<bool> = []` filled per body in `checkEffects`. This keeps `lspnavigation.nra`, which builds an `EffectState` literal, untouched.
- `narrac emit selfhost` from base and branch compilers: identical C (diff empty).
- `sh scripts/test.sh`: PASS (all shards, bootstrap job).
