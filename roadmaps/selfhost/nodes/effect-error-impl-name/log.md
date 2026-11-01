# Log: effect-error-impl-name

- Go: `userCall` in check_effects.go now passes `<Type>.<Trait>.<function>` as the "through" name for an implement block function (map built once per run from `traits.implFuncs`, type written with `Type.String()`, so no module prefix).
- Narra: `Callee` got a `label`; `findCallee` sets it (`describeImplementation` for impl functions); `walkUserCall` uses it.
- The types dump `func=` form carries the module (`basics.Cat.Named.name`); the error text follows the maintainer's form without it.
- New cases: `testdata/typesdump/effects/err_impl_through.{nra,dump}`, `testdata/errors/effects_impl_through.nra`.
