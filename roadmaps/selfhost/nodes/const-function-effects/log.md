# Log: const-function-effects

## Proof on feat/batch-21 (base compiler from `sh selfhost/bootstrap.sh`)

Program from the brief (`const shown = describe`, `describe` uses alloc):

    narrac check p1/src/main.nra
    main.nra:8:15: expected (int) => string, found (int) => string uses alloc

Re-export of an imported function with effects (`testdata/packages/const_function_effects`):

    narrac run testdata/packages/const_function_effects
    src/till.nra:3:34: expected (int) => string, found (int) => string uses alloc

Array literal of functions with effects (`testdata/const_function_effects.nra`):

    narrac run testdata/const_function_effects.nra
    const_function_effects.nra:19:10: expected Array<(int) => string uses alloc>, found Array<(int) => string>

`let`, struct fields and a ternary already worked: the store check follows the value.

## Cause

`checkFunctionValue` (selfhost/src/check/funcvalues.nra) built the type of a named function used as a value with an empty effect list. `checkConstType` (exprs.nra) then compared that type against the effects the effect pass finds for the value and reported the difference. The same empty list made a call through the constant lose the function's effects, and gave an array of such functions the element type without effects.

## Fix

`checkFunctionValue` puts the function's declared effects (not its effect variables) into the function type. A private function without a `uses` list keeps the old inference through the effect pass.

## Branch

    narrac check on the three programs above: clean; the const call now reports `Function report uses alloc but its signature does not declare it`
