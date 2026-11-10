# Outcome: int-to-string

## Built

- `selfhost/runtime/runtime.h`: `lstr_from_int` and `lstr_from_uint` share `lstr_from_digits`, which writes the digits backwards into a 24-byte stack buffer and copies them into one `lstr`. INT64_MIN is negated in unsigned arithmetic. No `snprintf`, no `strlen`.
- `selfhost/src/runtimecore.nra` regenerated (`--suite generated --update`).
- `testdata/int_to_string.nra` and `.out`: 0, negatives, powers of ten, min and max of `int`, `i8` to `i32`, and `u8` to `u64`. The `.out` was captured from the compiler before the change.

## Proof

Apple M2, `cc -O2 -DNARRA_RELEASE`, `/usr/bin/time -l <narrac> emit selfhost`, 4 alternating runs each, base feat/batch-4 `8bd8699`.

- base: 4.158, 4.160, 4.168, 4.200 G instructions retired; 139 MB peak.
- branch: 3.941, 3.941, 3.955, 3.957 G; 139 MB peak.
- Gain 5.3%; the spread on each side is under 1%.

## Deviations

- None. The seed is not regenerated.

## Stubs left

- None.

## Gotchas

- `Int.toString` takes `int` only; sized types print through `Show`, which uses the same functions.
- `narrac run` is the quick way to get a program sample output; `--update` does not write `programs` goldens.

## Verification

- `sh scripts/test.sh`: PASS in 94 s, 8 shards, 0 failed, including the sanitized bootstrap job.
- `sh selfhost/bootstrap.sh`: stage 2 and stage 3 emit identical C.

## Size

- Estimated S, actual S.
