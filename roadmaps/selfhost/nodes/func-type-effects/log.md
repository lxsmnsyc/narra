# Log: func-type-effects

- Design: `TFunc(parameters, result, effects, binder)`; type key compares effects and binder, C names do not. A plain type key (`plainTypeKey`) keys `Checker.composites`, so one C struct is emitted per C name.
- Values written as arrows or names list no effects in their type (the effect pass knows them only after the fixed point). `expectType` records a `FunctionStore` for a function value stored where a function type with the same signature is expected. The effect pass checks the stores after the fixed point, against what the value does (arrow set, named function, tracked local, else the effects of its type). Nested positions need equal types.
- Considered and dropped: one type per arrow with an origin slot (as generators). Two closures in one array would differ in type and every closure would instantiate a generic once.
- Private function, function-type parameter without `uses`: stays an implicit effect variable (`looseFunctions` while its arguments are checked). Builtin and std callbacks are loose too (their effects flow through `Builtin.Callbacks`).
- Assignment to a local is loose: the local tracks the union of what it holds (`walkLocalAssignment`), as before.
- Effect variables of a callee in a result type (`wrap(f)(1)`, `wrapAll(f)[0](1)`, `const fs = wrapAll(f)`): the pass finds the source call (`findSourceCall`, `EffectState.sources` for locals) and binds the variables from its arguments.
- Printing: `typeString` shows `uses ...` unless the variables belong to another function (keeps the std error goldens), and the C name table (comments in the emitted C) hides effects (`Checker.printEffects`).
- Review cases 14b (tuple index), 25 (never calls `h`) and 30 (`Arrays.get` does not exist) are invalid probes, not usable as tests; the others became `testdata/errors/func_effects_*` tests (43) plus the clean sample `testdata/func_effects.nra`.
- Fixed in selfhost: 4 struct field types in `fmtprinter.nra` (`uses alloc`). Test programs: `testdata/closures.nra` and `testdata/typesdump/closures/arrows.nra` (two function types each, `uses alloc`).
