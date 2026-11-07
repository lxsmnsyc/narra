# Outcome: effects-indirect-calls

## Built

- `compiler/internal/narra/check_effects.go`: `valueEffects` now handles struct fields, array and map elements and call results through the written type (`typeExprOf`, `writtenEffects`, `resultEffects`). A call through such a value charges the `uses` of its function type with the existing error messages.
- `selfhost/src/effects.nra`: the same rule (`findWrittenType`, `findWrittenEffects`, `findResultEffects`, `EffectState.localTypes`).
- Tests: `compiler/testdata/errors/effects_indirect_{field,index,call_result}.nra`, and the clean sample `compiler/testdata/effects_indirect.nra` (field, index, call result, local bound to a field, an effect variable bound to a field, and a pure field).
- No `uses` list was added to `selfhost/src`; it had no hidden indirect effects.

## Deviations

- None. The parenthesized case needs no code: the parser keeps no paren node, so `(box.run)(1)` is a field call.

## Stubs left

- None.

## Gotchas

- A local with no written type and a value with no written type (for-of variable, destructured name, match binding) is still treated as pure. Fixing that needs effects on `FuncType`.
- Effects on a field type that name an effect variable (`uses E`) are kept as written; struct types cannot declare effect variables.

## Decisions

- Follow the written type in the effect pass instead of putting effects on `FuncType`. It keeps type interning, equality and emitted C unchanged.
- Keys of struct lookups are module path plus name; tuple structs are skipped.

## Proposed learnings

- None.

## Verification

- See the PR description for the full check result.

## Size

- Estimated S, actual S.
