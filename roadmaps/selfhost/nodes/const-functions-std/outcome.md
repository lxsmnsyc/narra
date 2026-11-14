# Outcome: const-functions-std

## Built

- `@const` bodies call `std/math`, `std/bytes`, `std/fixedarray`, `std/f32` and `std/i8`, `std/i16`, `std/i32`, `std/u8`, `std/u16`, `std/u32`, `std/u64` (`CONST_STD_MODULES` in `selfhost/src/constfunctions.nra`). `std/char`, `std/option`, `std/result` and `std/regex` were already allowed and run (the sample covers them). `Math.sin`, `cos`, `tan`, `atan2`, `exp`, `log` and `pow` stay E1052: `checkCalledFunction` now tests the library list before the module list.
- `selfhost/src/conststd.nra` (new): the functions that run inside the compiler by calling the same std function: `I8.*` to `U64.*` (`from`, `wrap`, `tryFrom`, `toString`, `parse`, `addChecked`, `subChecked`, `mulChecked`; `u64` overflow tests compare without a sign), `F32.from`, `truncate`, `toString`, all of `Bytes.*`, `Libm.floor`, `ceil`, `round` and `trunc`, and `compareScalars` for `Ordered.compare` (numbers, `u64` unsigned, floats with the C rule `(a > b) - (a < b)`, chars, bools, strings, bytes). `wrapInteger` moved there from `consteval.nra`.
- `consteval.nra`: one `Ref.TraitFunction` row in `evaluateCall` (`callTraitFunction`, only `Ordered.compare`), one `callStdFunction` dispatch in `callBuiltin` and `finishOutcome`, which charges memory and steps and wraps the outcome as Some, None, Ok or Err.
- Source functions of `std/math` (`abs`, `min`, `max`, `clamp`, `powInt`, `sqrt`, `floor`...), `std/fixedarray` (`get`, `length`) run as Narra bodies, like the other source modules; none is done by hand. By hand (calls the std function inside the compiler, because the module is a builtin with no Narra body): the sized modules, `std/f32`, `std/bytes` and the four `Libm` functions.
- Bytes are `ConstValue.Items` of numbers. `emitconstvalues.nra` writes a bytes constant as a `static const` table copied into `lstr_new(n)` (core runtime, so it needs no `std_bytes.c`).
- Tests: `testdata/const_functions_std.nra` (+ `.out`): a constant for char, option, result, math (with `Ordered` on strings, floats, `u64` and bytes), bytes (also a `bytes` result constant), fixedarray, the seven sized modules and f32, each printed beside the same call at run time (every line has two equal halves). Error tests `testdata/errors/const_functions_std_assert`, `_index`, `_panic` (E1053 with call chain).
- SPEC section 12 item (modules list), `compiler/README.md` "Not supported yet", E1052 explanation in `diagnostics.nra`. The E1052 message lists `CONST_STD_MODULES`, so five existing `.err` goldens changed (longer module list).

## Proof

- base (feat/batch-15 at 37ffedff, release build): `narrac run testdata/const_functions_std.nra` gives five E1052 errors (`Math.powInt`, `Bytes.fromString`, `Fixed.get`, `I8.wrap`, `F32.truncate`: "a module the compiler does not run while it compiles").
- branch: the same command prints all lines with equal compile-time and run-time halves; `programs --filter const_functions_std` passes in debug (AddressSanitizer, leak and copy check) and release.

## Deviations

- Brief asks for an error test for "division by zero, overflow that traps". Neither can fail in an `@const` body: the checker proves every divisor (D92) and sized integers wrap. The three error tests use a failed assert, an index out of range after `Bytes` calls, and a `panic` after `Char` and `Results` calls.
- `F32.from` of an integer above 2^53 in size (or a `u64` from 2^63) is refused with E1052 at the constant: C converts it straight to float, and going through f64 could round twice.
- `Ordered.compare` through `min`, `max` and `clamp` runs for scalars, strings and bytes only; a struct or enum argument is E1052 at the constant ("the trait function Ordered.compare"). Other trait functions (`Equatable.equals`, `Hashable.hash`, `Show.show`) are not run.
- Files outside owns/shared: none.

## Stubs left

- `Result.map`, `andThen`, `mapErr`, `Options.map`, `andThen`, `orElse`, `Map.update` and `Arrays.map` take functions and stay out (D105 item 3).
- Ordering of struct and enum values in `min`/`max`.

## Gotchas

- A new builtin module with no Narra body needs rows in `conststd.nra` and a name in `CONST_STD_MODULES`; the scanner accepts it, the evaluator reports E1052 at the constant when no row runs it.
- A `from` clause is needed on functions that return a struct built from a parameter.

## Proposed learnings

- A std row the evaluator runs belongs in `conststd.nra`, with the checks the run-time function makes | consteval.nra is already 3800 lines | selfhost/src | no

## Verification

- Base and branch compilers emit identical C for `narrac emit selfhost` (`cmp`, 249565 lines on the branch source).
- `sh selfhost/bootstrap.sh`: fixed point (stage 2 and stage 3 identical), seed not regenerated.
- `sh scripts/test.sh` after the rebase onto feat/batch-15: `test.sh: PASS, 9886 cases in 203s`

## Size

- Estimated M, actual M.
