# Outcome: quote-on-error

## Built

- `useLocal` in `selfhost/src/funcvalues.nra` no longer calls `quoteGo(local.name)` on every use. It quotes the name inside each error message, so only the error paths pay for it. Messages are unchanged.

## Proof

- base: ffdea9c, `/usr/bin/time -l <narrac> emit selfhost`, 3 runs: instructions retired 9324375265, 9350551959, 9321813995; peak memory 127 to 137 MB.
- branch: same command, 3 runs: 9044939712, 9030800869, 9048181205; peak memory 136 to 138 MB.
- Gain about 3.0% (mean 9.332e9 to 9.041e9) against a spread of 0.3% on each side. Memory is unchanged within noise.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- The seed was not regenerated, as the brief says.

## Proposed learnings

- Build error text inside the error branch, not before the check | the string work ran on every use of a local | selfhost/src | no

## Verification

- `sh selfhost/bootstrap.sh <dir>`: stage 1 and stage 2 emit identical C.
- `narrac fmt selfhost/src/funcvalues.nra`: no change.
- `sh scripts/test.sh`: see the report.

## Size

- Estimated S, actual S.
