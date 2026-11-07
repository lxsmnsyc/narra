# Outcome: effects-indirect-calls

## Built

- `compiler/internal/narra/check_effects.go`: `valueEffects` handles struct fields, array elements and call results through the written type (`typeExprOf`, `writtenEffects`). A call through such a value charges the `uses` of its function type with the existing error messages. `bindUses` replaces the effect variables of a callee in its result type by the effects bound at the call, on every path.
- `selfhost/src/effects.nra`: the same rule (`findWrittenType`, `findWrittenEffects`, `bindUses`, `EffectState.localTypes`).
- Effect names in struct field function types are now checked like those in signatures; an effect variable there is rejected (a struct declares none).
- Tests: `compiler/testdata/errors/effects_indirect_{field,index,call_result,capture,paren,field_unknown,field_var}.nra`, and the clean sample `compiler/testdata/effects_indirect.nra` (field, index, call result, local bound to a field, an effect variable bound to a field and to a call result, parenthesized callee, pure field).
- No `uses` list was added to `selfhost/src`; it had no hidden indirect effects.

## Deviations

- None. A parenthesized callee needs no code: the parser keeps no paren node.

## Stubs left

- None.

## Gotchas

- The effect pass trusts the written type of a value but does not enforce it. Holes that remain, all accepted on main as well:
  - Stores: a function value put into a field, array, return slot or struct literal whose type has more effects than the value is not checked.
  - Type parameters erase effects: `id(f)(1)`, `first(fs)(1)`, `Options.withDefault`, generic struct fields.
  - Closures that return a field, match expressions yielding a field, and arrow parameters.
  - A value with no written type (for-of variable, destructured name, match binding) counts as pure.
- The real fix is effects on `FuncType`, the counterpart of D78 for generators. The orchestrator is adding it as a separate roadmap node. This PR is an improvement over main, which accepts every field, index and call-result call.
- Map elements are not followed: `m[k]` is `Option<V>`.

## Decisions

- Follow the written type in the effect pass instead of putting effects on `FuncType`. It keeps type interning, equality and emitted C unchanged.
- Substitute the bound effects into the callee's result type (`bindUses`) instead of giving up when a result names an effect variable.
- Struct lookups are keyed by module path and name; tuple structs are skipped.
- The unknown-effect error for a struct field reuses the signature message, so an effect variable there reads "unknown effect `E`; ... or an effect variable declared as `<effect E>`".

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test -timeout 45m ./...`: see the PR comment for the result line.

## Size

- Estimated S, actual S.
