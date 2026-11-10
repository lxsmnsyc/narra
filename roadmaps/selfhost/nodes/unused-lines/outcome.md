# Outcome: unused-lines

## Built

- `selfhost/src/unused.nra`: `reportUnusedImports` and `reportUnusedFunctions` split the module source into lines on the first unused name, not at the start. Diagnostics and fixes are the same.

## Proof

- base: 6924411, `narrac check selfhost`, 20 interleaved runs: median 0.407 s, mad 0.002, range 0.403 to 0.453.
- branch: same command: median 0.382 s, mad 0.002, range 0.376 to 0.389. Gain 25 ms (6.1%), far above the noise. Method and the noisy first look are in `log.md`.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- Timing on a loaded machine varied 25% between runs; use interleaved runs and the median.
- `reportUnusedImports` still walks each module's imports; only the split is gone.

## Proposed learnings

- Do not split or copy a module source before knowing there is something to report | the split was 6% of a check | selfhost/src | no

## Verification

- `sh scripts/test.sh`: PASS, all shards 0 failed (75 s).
- `narrac fmt selfhost/src/unused.nra` run.

## Size

- Estimated S, actual S.
