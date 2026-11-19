# Outcome: const-float-exact

## Built

- `selfhost/src/check/consteval.nra`: `parseFloatText` tries the exact short path (`parseShortFloatText`), then `parseLongFloatText`, which rewrites the literal as digits and a power of ten and reads it with `std/json` (C library reader, correctly rounded). Infinite results are still rejected.
- `testdata/const_float_literals.nra` and `.out`: `0.49999999999999994`, `1e300`, `5e-324` and a 36-digit literal, as call arguments and in function bodies; each folded value is compared with the run-time value (`==`) and printed.

## Proof

- base: 94107ca4, `NARRAC=<base narrac> sh scripts/test.sh --filter const_float_literals`: FAIL, "constant HALF_BELOW cannot be computed while compiling: it uses the float literal 0.49999999999999994, ...".
- branch: same command: pass; `--filter const_` passes 459 cases.

## Deviations

- The `F32.from(int)` beyond 2^53 item is skipped: it needs a direct int-to-f32 conversion in the evaluator and the emitter, more than a small change.

## Stubs left

- None.

## Gotchas

- The f32 tie check still uses `isExactDecimal`, so a long f32 literal that is a tie between two f32 values still stops evaluation with E1052.
- `consteval` now imports `std/json`, so the compiler's `@const` evaluation depends on it.

## Proposed learnings

- Do not edit `selfhost/src` while `selfhost/bootstrap.sh` runs | stage 2 reads the edited source and the fixed point check breaks | selfhost | no

## Verification

- Full `sh scripts/test.sh`: see the result line in the report.

## Size

- Estimated S, actual S.
