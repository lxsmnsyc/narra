# Outcome: sized-f32

## Built

- `Type.TF32` (`types.nra`), C `float`, type id `f32`; `"f32"` registered in `newScope` and `stdType` (`check.nra`, `stdtypes.nra`). `isNumeric` accepts it.
- Literals (`exprs.nra`): `checkFloatLiteral` gives `f32` when the wanted type is `f32`, else `f64`; `f32LiteralFits` reads the digits and exponent and compares with 3.4028235677973366e38 (E0297). A wanted `f32` types the left operand of an arithmetic operator; a float literal on the left takes the type of the right operand (`isNumberLiteral`). Emit adds an `f` suffix (`floatLiteralCode`, `emitexprs.nra`); `%` on `f32` is `fmodf`.
- `f32` against `f64` in any binary operator except `&&`/`||` is E0295 with a float message (second row of E0295 in `diagnostics.nra`). `f32` with `int` stays E0217/E0214/E0215 as `f64` with `int`.
- Traits: `Equatable`, `Ordered`, `Show` through `narra_eq_f32`, `narra_cmp_f32`, `lstr_from_f32` (`runtime.h`; shortest text that `strtof` reads back). Not `Hashable`; map key problem text "float keys are not allowed". Literal patterns accept `f32` (`match.nra`).
- `selfhost/src/stdf32.nra`: `F32.from` (f64 or any integer), `F32.truncate`, `F32.toString` (module `std/f32`) and `Float.from` (f32 or any integer, module `std/float`). `from` is a C cast. Wrong argument types are E0941 (second row, "takes %s or an integer"). Registered in `stdlib.nra`, `stdcalls.nra`, `emitstd.nra`.
- SPEC section 4 "Numbers", `compiler/README.md`. `runtimecore.nra` regenerated. `unknown_module.err` golden updated (module list).
- Tests: `testdata/f32.nra`/`.out`; `testdata/errors/f32_{range,mixed,key,argument}`.

## New codes

- E0297 `float literal %s does not fit in f32`. Also extra rows (no new numbers): E0295 float message, E0941 "takes %s or an integer".

## Gate decisions

1. Range test is on the rounded result: a literal at or above 2^128 - 2^103 (3.4028235677973366e38) is E0297; smaller ones round to a finite f32. Alternative: reject above `FLT_MAX` (3.4028235e38) only. Rejected: it rejects literals that round to `FLT_MAX`. Tiny literals that round to zero are accepted.
2. Mixed `f32`/`f64` reports E0295 (as for integers) with its own message. Alternative: E0217. One rule, one code.
3. `Float.from` accepts `f32` and integers only, not `f64` (identity); `F32.from` accepts `f64` and integers, not `f32`. Other types E0941.
4. No DECISIONS entry (none reserved); the rules are in SPEC.

## For dependents

- `Type.TF32` is a new variant; exhaustive matches on `Type` need a case. `Float.truncate` panic text is shared by `F32.truncate` ("Float.truncate: value out of range").
- `Float.toString` takes `f64` only. `F32.toString` is the string form.
- The seed is not regenerated.

## Verification

See the report (full `scripts/test.sh`).
