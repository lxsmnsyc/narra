# Outcome: sized-conversions

## Built

- `selfhost/src/stdsized.nra`: modules `std/i8 i16 i32 u8 u16 u32 u64` (aliases `I8` ... `U64` by convention) with `from`, `tryFrom`, `wrap`, `toString`, `parse`; `Int.from`, `Int.tryFrom`, `Int.wrap` registered from the same table (`stdlib.nra`) and emitted through `emitIntCall`.
- `from` and `wrap` emit a C cast; `tryFrom` a per-pair range check returning `Option`; `toString` calls `lstr_from_int` or `lstr_from_uint`; `parse` calls `narra_sized_parse` (new, `std_int.c`), error text `number does not fit in 8 bits (-128 to 127)`; other errors as `Int.parse`.
- `stdType` accepts the sized type names (`stdtypes.nra`). Registration after `std/test`, so the unknown-module message now lists the seven modules last (goldens updated).
- SPEC section 4 and the Standard modules list; `compiler/README.md` "Supported".
- Tests: `testdata/sized_convert.nra`/`.out` (every lossless pair, tryFrom boundaries, wrap, toString and parse limits; ASan clean), `testdata/errors/sized_convert_{lossy,argument}`.

## New codes

- E0940 `X.from cannot take T without loss; use X.tryFrom or X.wrap`; E0941 `X.fn takes an integer, found T`. Reported at the argument.

## Gate decisions

1. Literal argument of `from` is an `int` (no hint), so `I32.from(5)` is E0940. Alternative: type a literal from the target. Rejected: it hides which width the literal has and breaks `I32.from(1 + small)`. `tryFrom` and `wrap` take literals as `int`.
2. `parse` out of range: one message with the range, also for a minus sign on unsigned. Alternative: separate "negative" text.
3. E0941 applies to `tryFrom` and `wrap` too.
4. No DECISIONS entry (none reserved); D82 already names the functions.

## For dependents

- `Show.show` remains for templates of generic code; `X.toString(v)` is the string form. Std `toString` takes only its own type.
- `reportMixedIntegers` names match; unchanged.
- The seed is not regenerated; the comment in `exprs.nra` saying the functions come later is stale.

## Verification

- `NARRAC=out/narrac sh scripts/test.sh --update`: PASS (4782 cases). Only the unknown-module goldens changed.
