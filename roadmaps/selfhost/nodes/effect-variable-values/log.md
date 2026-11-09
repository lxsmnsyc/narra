# Log: effect-variable-values

## Proof on the base

- Base: `6478acd` (origin/feat/batch-2). Compiler: `sh selfhost/bootstrap.sh <dir>` from the seed.
- Command: `narrac build -o out testdata/errors/effect_variable_value_top.nra` and `..._local.nra`. Each has a pure exported `pure()` that calls `apply(loud, 1)`, with `loud` doing `io` and `apply` being `twice` (an effect-variable function) in a top-level constant, or in a local constant.
- Result on the base: both exit 0 with no diagnostic. Reproduced.

## Cause

`apply(loud, 1)` is a call through a value. `valueEffects` gives the tracked set of the local (`concreteEffects` of `twice`, which drops the variable `E`) or the type effects of the constant, and nothing binds `E` to the arguments.

## Choice

Option 1. The substitution of `findBoundEffects` already binds a variable from the arguments; the value's type (`TFunc` with its binder) holds the parameter types that name the variable, so the call can bind them without new syntax.
