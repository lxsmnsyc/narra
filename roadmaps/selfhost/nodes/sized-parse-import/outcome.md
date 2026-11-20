# Outcome: sized-parse-import

## Built

- `selfhost/runtime/std_sized.c` holds `narra_sized_parse` (moved out of `std_int.c`); generated `selfhost/src/runtime/runtimestdsized.nra`; `runtimeFor("std/internal/sized")` returns it.
- `emitParse` in `selfhost/src/std/stdsized.nra` notes the helper `std/internal/sized`; `emitProgram` in `selfhost/src/emit/emit.nra` writes its runtime when noted.
- Sample `testdata/sized_parse_alone.nra` (+ `.out`): parse in every sized module, toString, f32, no std/int import.

## Proof

- base: feat/batch-21, `narrac run testdata/sized_parse_alone.nra` and the brief's program: C compiler error `call to undeclared function 'narra_sized_parse'`.
- branch: same command prints the expected output; the brief's program prints `ok`.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- The sized runtime is keyed by a helper note, not by `checker.used`, because no module path names it.

## Proposed learnings

- A runtime function used by a module with no runtime file goes in its own `std_NAME.c` and is noted with `noteHelper` | a function in another module's runtime file only exists when that module is imported | selfhost/runtime | no

## Verification

- `sh selfhost/bootstrap.sh`: fixed point reached.
- `sh scripts/test.sh`: PASS, 10695 cases.

## Size

- Estimated S, actual S.
