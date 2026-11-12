# Outcome: std-function-values

## Built

- `selfhost/src/stdvalues.nra`: `checkStdFunctionValue` (function type from the registry row, `Ref.Builtin` on the node, E1013 otherwise) and `emitStdFunctionValue` (one thunk per builtin, a synthetic call with negative node ids run through `EmitWalker.emitExpression`).
- Dispatch only in `exprs.nra` (`checkField`, `checkName`), `effects.nra` (`findFunctionValue` reads the row's effects), `emitexprs.nra` (`emitIdent`, `emitFieldRead`).
- Emitter helpers in `emitstate.nra`: `findSyntheticCode` (the side-table lookup, also read by `readNodeType` and `readNodeRef`), `enterBodyContext` and `leaveBodyContext` (save and restore of the per-body state). node inline reuses these. `emitClosureFunction` uses the pair too.
- E1013 row and explanation in `diagnostics.nra`. SPEC section 2 "Functions" and `compiler/README.md` updated.
- Samples: `testdata/std_function_values.nra`. Errors: `testdata/errors/std_function_value_{generic,mutable,effects,type}`.

## Proof

Not a fix node.

## Deviations

- `checkStdFunctionValue` takes `shown` (the name as written) and no `want`: the hint must say `Arrays.length`, while the registry name is `Array.length`, and a fixed signature needs no expected type.
- The emitter hook did not spread: one thunk builder covers every eligible row.
- `Int.toString` has `alloc`, so function types receiving it say `uses alloc` (the SPEC example for `Fs.readText` already lists effects).
- Prelude functions (`print`, `assert`, `panic`) are not values.
- `testdata/errors/prelude_imported_value.{nra,err}` (not owned) changed: `const f = Arrays.length` now reports E1013 instead of "a module is not a value".

## Stubs left

- None.

## Gotchas

- Synthetic node ids are negative (-2 and below); `findBodySlot` returns -1 for them, so any reader that skips `readNodeType`/`readNodeRef` sees nothing.
- Unlike the brief's wording, the helper is three functions (lookup, enter, leave), because Narra function types cannot take `mutable`.

## Proposed learnings

- A builtin emitter must read argument types through `readNodeType` and refs through `readNodeRef` | synthetic nodes only exist in those tables | selfhost/src | no

## Verification

- `sh scripts/test.sh`: `test.sh: PASS, 7200 cases in 180s`.
- Samples ran under ASan/UBSan and `NARRA_CHECK_LEAKS=1`, clean.

## Size

- Estimated M, actual M.
