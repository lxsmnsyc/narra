# Outcome: map-helpers-split

## Built

- `selfhost/src/emitmaps.nra`: `emitMapFunctions` splits the map helper template (`RUNTIME_MAP_H`) on `@` once into a `MapTemplate` (`texts`, `slots`) with `splitMapTemplate`. `mapFunctions` takes the template and fills the 12 placeholders of `MAP_PLACEHOLDERS` in one pass, with no `Str.replace`. A piece between `@` marks that names no placeholder (the `@...@` in the template's header comment) stays text.

## Proof

- base: commit 43aae732 (`feat/batch-5`), `/usr/bin/time -l <narrac> emit selfhost`, instructions retired, 8 alternating pairs, load average about 14. Runs: 1.919 G to 1.950 G, mean 1.936 G, spread about 1.7%.
- branch: same command, runs 1.820 G to 1.841 G, mean 1.831 G. Won 8 of 8 pairs, 5.4% fewer instructions, well above the spread.

## Deviations

- The gain (5.4%) is larger than the profile's estimate (about 2%): each `Str.replace` also copied the 4 KB text, which the sample share did not count.

## Stubs left

- None.

## Gotchas

- The seed was not regenerated; `bootstrap.sh` notes that the seed is older than the source, as expected during the batch.
- The placeholder order in `MAP_PLACEHOLDERS` must match the order of `values` in `mapFunctions`. `releaseCode`, `retainCode`, `useEq` and `useHash` register needs, so the values are computed in the old order.

## Proposed learnings

- None.

## Verification

- `emit selfhost` output is byte-identical on base and branch (`cmp`); `emit` of 61 testdata samples that use `Map<` is identical too.
- `narrac fmt selfhost/src/emitmaps.nra` applied.
- `NARRAC=<branch narrac> sh scripts/test.sh`: PASS (8 shards, 5,858 cases, plus the sanitized bootstrap job). Self-review, one round, no findings.

## Size

- Estimated S, actual S.
