# Outcome: effect-int-keys

Done. `selfhost/src/effects.nra` keys `bodyIndex`, `functions`, `locals`, `sources`, `entries` and `valuesAt` by integers.

- `packPosition(owner, pos)` packs a module scope index (or a body index) with line and column; `packLocal` packs body and local index. Callee keys come from `functionKeys` (one per Checker.functions entry) and `bodyKeys` (one per body), built once in `checkEffects`.
- `valuesAt` holds a slot into `valueLists`, and the push goes to `state.valueLists[slot]`, so the array is no longer copied on each push.
- The string key remains only for `checker.generatorEffects` (owned by the checker), built for private generators only.

## Evidence

`/usr/bin/time -l narrac emit selfhost`, 3 runs each:

| | instructions retired | peak memory |
|---|---|---|
| base (ffdea9c) | 9.39e9, 9.36e9, 9.35e9 | 121-134 MB |
| branch | 8.37e9, 8.29e9, 8.32e9 | 115-123 MB |

About 11% fewer instructions; the spread is under 0.5%. The bootstrap fixed point holds. `sh scripts/test.sh` passes (all shards).

## Deviations

None.
