# Outcome: const-function-effects

## Built

- `selfhost/src/check/funcvalues.nra`: `checkFunctionValue` gives the type of a named function used as a value the effects the function declares (`listDeclaredEffects`, effect variables left out). Before, the list was empty.
- A constant with no written type, `export const` re-exports, array literals of such functions and calls through the constant now see those effects.
- Cases: `testdata/const_function_effects.nra` (constant, `let`, struct field, array literal), `testdata/packages/const_function_effects` (re-export of an imported function), `testdata/errors/const_function_effects_call.nra` (a caller without `uses alloc` is reported).
- `testdata/packages/errors/effects_const_import/expected.err` lost its line for `export const step = loud`, which was the bug.
- Chapter 4, "Export a name from another module": the `shop` sample re-exports `describe`, which uses `alloc`.

## Proof

- base: feat/batch-21, `narrac check` of the brief's program: `expected (int) => string, found (int) => string uses alloc` at `const shown = describe`. The package case fails the same way at `src/till.nra:3:34`, the array case with `expected Array<(int) => string uses alloc>, found Array<(int) => string>`.
- branch: the same commands are clean. A caller of the constant without `uses alloc` gets `Function report uses alloc but its signature does not declare it`, which the base never reached.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- An array literal mixing a function without effects first and one with effects (`[quiet, describe]`) still fails, as does `cond ? describe : quiet`: the element type is taken from the first. A written type fixes it.
- A private function with no `uses` list that allocs is still not accepted as the value of an untyped constant (its effects are known only after the effect pass); write the constant's type or the function's `uses` list.
- A call through a constant was not checked for effects before; callers in other code may now report missing `uses`.

## Proposed learnings

- None.

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh --suite guide`: PASS, 601 cases (guide suite).
- `sh scripts/test.sh`: PASS, 10787 cases in 233s.
- Two existing cases (`testdata/errors/effects_const_function`, `effects_const_copy`) expected the const-site error; they now expect the missing-`uses` error alone.

## Size

- Estimated S, actual S.
