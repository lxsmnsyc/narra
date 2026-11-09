# Log: generator-function-type

## Proof on the base

- Base: 6478acd (origin/feat/batch-2), narrac built with `sh selfhost/bootstrap.sh`.
- Command: `narrac run testdata/generator_function_type.nra` (top-level and local annotated bindings, and a call argument for a parameter of the annotated type).
- Result: reproduced. All three sites fail with `expected (int) => Generator<int>, found (int) => Generator<int>`.

## Causes found

1. `expectType` compared function results with `==`. The result of a private `function*` carries its origin (the function index), the annotation has origin -1, so the types differ though they print alike.
2. Creating a generator charges `alloc` to the function value. A function type cannot list that effect apart from its generator result (`uses` after `Generator<T>` binds to the generator), so the store check rejected it.
3. The C names of function types (`F_Fn1_...`) embedded the generator effects and origin, so the function* and the annotation got two different C structs. The generator C type is always `narra_gen*`.

## Fix

- `exprs.nra`: a function type whose results are generators of the same element and input is accepted; the generator effects are checked with `expectGenerator`.
- `effects.nra`: the `alloc` of creating a generator is ignored when the stored-into type returns a generator.
- `cnames.nra`: a generator id is its element and input only.
- `emitclosures.nra`: function types that share a C name are listed once.
