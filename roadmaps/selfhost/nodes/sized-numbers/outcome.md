# Outcome: sized-numbers

## Built

- Types `i8 i16 i32 u8 u16 u32 u64` (`Type.TI8` ... `Type.TU64` in `types.nra`; `i64` stays `int`). Helpers `isInteger`, `isSizedInteger`, `isUnsignedInteger` in `types.nra`. C names `int8_t` ... `uint64_t`; type ids `i8` ... `u64`.
- Literals (`exprs.nra`): `checkIntLiteral` takes the wanted sized type, else `int`, and checks the range (E0294). A literal directly under unary minus is checked with its sign (`Checker.negatedLiteralId`), so `-128` is an `i8`. u64 range is compared by digit count, since `int` cannot hold it.
- `checkBinary` takes `want`: a sized wanted type types the left operand, and a literal on the left takes the type of the right operand. Different integer types (one sized) are E0295 for every binary operator except `&&` and `||`. Bitwise operators and shifts need two integers of one type. `-x` on an unsigned non-literal is E0296. `~` accepts any integer. Literal patterns accept any integer type (`match.nra`).
- Emit: sized literals are cast (`((uint8_t)INT64_C(200))`); u64 literals above `int`'s maximum are wrapped into int64 then cast. Arithmetic calls `narra_add_u8` etc. (`emitexprs.nra`). Eq, hash, cmp and show use `narra_<op>_<type>` (`emittraits.nra`).
- Runtime: the `NARRA_SIZED` macro in `runtime.h` defines add, sub, mul, neg, div, rem, shl, shr, wrapping add, sub, mul, eq, cmp, hash and show per type; `lstr_from_uint` prints unsigned. `runtimetext.nra` regenerated.
- Diagnostics E0294, E0295, E0296 in `diagnostics.nra`. SPEC section 4 "Numbers" and DECISIONS D82 written.
- Tests: `testdata/sized_int.nra` and `.out`; `testdata/errors/sized_int_{range,mixed,negate,argument}.nra` and `.err` (E0294, E0295, E0296, E0119); `testdata/panics/sized_int_overflow.nra` plus a line in `selfhost/tests/src/panics.nra`.

## Deviations

- The brief asks for a `std/test` panics-form test of `i8` overflow. A `Test.panics` sample would also run in release mode in the `programs` suite, where overflow wraps and does not panic. The `panics` suite (debug build, exit code 101, message `integer overflow`) covers it instead.
- A negative literal under an unsigned type (`const y: u8 = -1`) reports E0294, as rule 2 says; E0296 is only for `-x` on an unsigned value that is not a literal.
- `stage1` and `stage2` C differ until the seed is regenerated (the seed's runtime text is older). Stage 2 and stage 3 are identical. `sh selfhost/bootstrap.sh` reports FAIL on this until the orchestrator regenerates the seed.

## Gate decisions

All recorded in D82 (reserved number). Chosen for the maintainer by the orchestrator's rule for this batch:

1. Negative literal under an unsigned type: E0294 (alternative: E0296). The literal is the mistake and the message gives the value.
2. Mixed integer types report one new code E0295 for arithmetic, bitwise, shift and comparison, including `==` (alternative: keep E0214/E0215 for comparisons). One rule, one code, as rule 3 says.
3. Left operand typed from the wanted type, and a left literal typed from the right operand (alternative: left operand always checked with no hint, so `1 + d` fails). Needed for rule 2 ("type the literal from the other side").
4. Shift count has the type of the shifted value (alternative: `int` count). Keeps rule 3 literal; a count at or above the width panics.
5. Printing: `Show.show(x)`; template parts stay strings (E0201). The brief's "prints in template strings" is met by sized-conversions' `toString`.
6. u8 to u32 hash through the `int` hash so equal values hash alike.

## For sized-conversions

- The types exist and all trait functions work; add the std modules `std/i8` ... `std/u64` and `from`, `tryFrom`, `wrap`, `toString`, `parse` to `std/int`. E0295's message already names `<Module>.from`, `.tryFrom`, `.wrap` where `<Module>` is `I8 I16 I32 U8 U16 U32 U64`, and `Int` for `int`. Change `reportMixedIntegers` in `exprs.nra` if the names differ.
- Add a std builtin signature for each function in `stdlib.nra` (it holds `stdType` for the primitive names `int f64 bool char string void StringBuilder`; add the sized names to its `prims`/`types` lists there so std signatures can mention them).
- Runtime functions you can reuse: `lstr_from_int`, `lstr_from_uint` (unsigned text), `narra_show_<type>`. `wrap` is a plain C cast to the target type; `tryFrom` needs a range check per pair.
- `typeId` and `cType` give the C names (`cnames.nra`).
- `Int.toString` takes `int` only (an `i32` argument is E0119).

## For sized-f32

- `f32` can follow the same pattern: add `Type.TF32`, register `"f32"` in `newScope` (`check.nra`) and `typeString`, `typeKey`, `typeId`, `cType` (all list the primitive variants without a wildcard, so the compiler points at each). `isNumeric` in `check.nra` is where a float joins arithmetic; `checkIntLiteral` shows how a context type picks a literal's type (a float literal would do the same). The trait functions for primitives are in `emittraits.nra` (`isPrimitive`, `primitiveName`) and the `NARRA_SIZED` macro in `runtime.h` is the pattern for per-type functions.

## For bytes-type

- `Array<u8>` stores one byte per element (it is `uint8_t*` data in the generated array struct). A `bytes` type can build on `u8`: indexing returns `u8`, and integer literals typed `u8` are checked in range.

## Stubs left

- None. No std conversion functions (sized-conversions), no `f32`.

## Gotchas

- `sh selfhost/bootstrap.sh` takes minutes when other workers load the machine. To test without the fixed point, build stage 2 by hand: `narrac0` (cc seed), `narrac0 emit selfhost`, cc, then `narrac1 emit selfhost`, cc. Run the suites with `NARRAC` set to that.
- Unary minus on an `IntLit` is the only place that needs `negatedLiteralId`; it is reset after the operand check.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh` with `NARRAC` set to the stage 2 compiler: see the report.
