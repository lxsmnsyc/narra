# Outcome: map-key-panic-sample

## Built

- `testdata/panics/map_missing_key.nra`: `counts[key] += 1` on a missing `Map<string, int>` key.
- A `PanicSample` row in `selfhost/tests/src/panics.nra` expecting "map key not found" (the text in `emitstmts.nra`).

## Deviations

- The brief owns `testdata/panics/map_missing_key*`, which matches the real location. None otherwise.

## Stubs left

- None.

## Gotchas

- A read `counts["a"]` has type `Option<int>`, so the sample returns 0.

## Proposed learnings

- None.

## Verification

- `NARRAC=out/narrac sh scripts/test.sh --suite panics`: PASS (9 samples).
- `sh scripts/test.sh`: see report.

## Size

- Estimated S, actual S.
