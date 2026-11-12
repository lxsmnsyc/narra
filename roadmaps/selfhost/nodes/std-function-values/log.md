# Log: std-function-values

- Branch feat/std-function-values reset to feat/batch-11 (20f9a487).
- Checker: `checkStdFunctionValue` in `selfhost/src/stdvalues.nra`, called from `checkField` (alias of a std module) and `checkName` (imported name). The hint uses the name as written (`Arrays.length`, not the registry name `Array.length`).
- Effects: `findFunctionValue` reads the row's effects for `Ref.Builtin`. Passing a wider-effect value to a narrower type behaves as for user functions (the storing error and the "Function uses" error).
- Emitter: `emitStdFunctionValue` writes `thunk_std_<Module>_<name>` once. The thunk builds a synthetic `CallExpr` with negative node ids and runs it through `EmitWalker.emitExpression`, so every builtin emitter works unchanged. Types, refs and C text of the synthetic nodes live in three Emitter maps, read first by `readNodeType`, `readNodeRef` and `emitIdent`.
- Helpers in emitstate.nra: `findSyntheticCode` (the lookup), `enterBodyContext` and `leaveBodyContext` (save and restore of the per-body state). `emitClosureFunction` now uses the pair too.
- Row eligibility: check is "" or "fixed", no `mutables`, no `callbacks`. Everything else is E1013. Prelude functions (`print`, `assert`, `panic`) are not covered.
- `Int.toString` has the `alloc` effect, so a function type that receives it must say `uses alloc`.
