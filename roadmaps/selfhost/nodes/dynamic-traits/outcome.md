# Outcome: dynamic-traits

## Built

- `Type.TDynamic(traitIndex)` (`selfhost/src/types.nra`), resolved from `Dynamic<Trait>` in `resolveDynamicType` (`resolve.nra`, next to `Box`); a user type, generic or alias named `Dynamic` hides it. `typeString` prints `Dynamic<Drawable>`, `typeKey` is `Q<index>`, the C type is `struct narra_dyn*` and the C id `Q_<module hash>_<Trait>` (`cnames.nra`).
- `selfhost/src/dynamic.nra`: `flushDynamicTraits` checks each written `Dynamic<Trait>` once the trait's functions are resolved (E1023, naming the function). `isDynamicName` and `checkDynamicOf` check `Dynamic.of(value)` (E1024 without an expected `Dynamic<Trait>`, with another argument count or another function name; E0524, or E0523 for a type parameter, when the value's type does not implement the trait) and make the instances its table calls through `findImplementation`. `checkDynamicCall` reports E1025 for a `mutable Self` function and records the Dynamic in `traitSelf`. Dispatch: one `if` in `checkDottedCall` and one in `checkTraitCall` (`calls.nra`).
- A `Dynamic` implements no trait (`computeImplements`), is never a map key (`mapKeyProblem`: "a Dynamic is not hashable"), and is counted (`isCountedType`). The shared-result checks follow a trait function's `from` through a Dynamic (`findSharedArgument`, `canShareType`, `canShareKnown`, `searchHeapValue`), as its stored value may hold a tracked value.
- `selfhost/src/emitdynamic.nra`: one cell type per (trait, type) pair, `{ narra_dyn header; T value; }`, with a drop function, one thunk per trait function (a user block, an instance of a generic block through `lookupImplementation`, or the compiler's `Show` and `Hashable` functions through `useShow` and `useHash`), a static const table whose first member is the drop function, and `<cell>_of`. `narra_dyn`, `narra_dyn_retain` and `narra_dyn_release` are written after the runtime when emitted code uses a Dynamic, so the runtime text did not change. `Drawable.draw(item, ...)` emits `((const struct Q_..._table*)item->table)->f_draw(item, ...)`. Dispatch: `Ref.Unresolved` in `emitCall` (`emitexprs.nra`), the cell loop and the text sections in `emit.nra`, `TDynamic` in `refCountCode` (`emitstate.nra`).
- Samples: `testdata/dynamic_traits.nra` (array loop, parameter, result, struct field, `Option`, extra parameters with a `mutable Canvas`, a large struct passed by pointer, copies outliving the original, copies pushed and dropped in a loop, a map of Dynamics), `testdata/dynamic_traits_show.nra` (`Dynamic<Show>` of int, string, char, bool, an array, an `Option<f64>`, a derived struct and enum, and a user `Show`), `testdata/dynamic_traits_generic.nra` (`Dynamic.of(value)` in `<T: Label>` bodies; `Dynamic<Label>` and `Dynamic<Show>` over `Pair<T>` with generic implement blocks, nested), `testdata/dynamic_traits_hash.nra` (`Hashable.hash` through a `Dynamic<Hashable>`), `testdata/packages/dynamic_traits` (trait from another module). Errors: `testdata/errors/dynamic_trait_{argument,rule,of,mutable,equality,key,shared,shared_result}.nra`. Query: `testdata/query/dynamic_traits` (hover and interface show `Dynamic<Drawable>`).
- E1022 to E1025 rows and explanations in `diagnostics.nra`. SPEC "Trait values" states the trait rule, the layout and the D81 rule; README drops the "Not supported yet" line and the D81 half-line and amends "every trait call is a direct call"; D95 has a "Dynamic traits" built paragraph with the amendments.

## Deviations

- Calls through a Dynamic and `Dynamic.of` keep `Ref.Unresolved`, not `Ref.TraitFunction`: the effect pass already gives an unresolved trait call the trait function's declared effects (as for a bound), and the shared-result pass treats it as a bound call. traitSelf holds the Dynamic for a call and the stored type for `Dynamic.of`.
- Files outside owned and shared paths: `selfhost/src/body.nra` and `selfhost/src/declare.nra` (one `flushDynamicTraits` call and import each, beside `flushTypeBounds`), `selfhost/src/effects.nra` (`Dynamic.of` adds `alloc`, 4 lines in `addTraitCallEffects`), `testdata/packages/dynamic_traits/**` and `testdata/query/dynamic_traits/**` (the package sample and the hover test the acceptance asks for).
- `copy(d)` is E0975 (copies nothing), not a retain: a Dynamic is never changed in place, like a `string`. Recorded in D95.
- `Dynamic<Hashable>` is allowed (orchestrator, after the fresh review), against the brief's list. Recorded in D95.
- `checkTraitCall` no longer reports E0522 when the first argument's type is void, so a failed `Dynamic.of` does not cascade (review nit). This drops one cascading E0522 from `testdata/typesdump/calls/err_infer.dump` (outside the owned globs).
- The trait rule is queued (`Checker.dynamicTraits`, `DynamicTraitUse` in check.nra) because a field `Dynamic<Drawable>` is resolved before the trait's functions are.

## Stubs left

- None.

## Gotchas

- The table holds no slot for a `mutable Self` function, since E1025 rejects every call of one through a Dynamic.
- A call through a Dynamic is `Ref.Unresolved` with `traitSelf[nodeId]` set. Code that treats `Ref.Unresolved` as "an open generic call" must check traitSelf first.
- `struct narra_dyn*` is the C type everywhere; the `narra_dyn` typedef exists only when the "dynamic" helper was noted. Thunks and cells use the typedef, and noting happens whenever a cell or a table call is emitted.
- A `Dynamic` does not meet bounds (`<T: Show>`), so `Show.show(items)` on an `Array<Dynamic<Show>>` is rejected; the D95 amendment says why.

## Proposed learnings

- A new counted Type variant must be listed in `isCountedType` (check.nra) | otherwise locals of the type are never released and the leak check is the only signal | selfhost/src | yes (a sample with a local of the type under the leak check)

## Verification

- After the review fixes and the rebase: `--suite errors`: `PASS, 666 cases`.
- `NARRAC=<dev narrac> sh scripts/test.sh --suite programs --filter dynamic`: `test.sh: PASS, 6 cases` (debug with sanitizers and release, both with NARRA_CHECK_LEAKS=1 and NARRA_CHECK_COPIES=1).
- `--suite errors --filter dynamic_trait`: `PASS, 6 cases`; `--suite query`: `PASS, 7 cases`; `--suite packages`: `PASS, 36 cases`; `--suite codes`: `PASS, 1 cases`.
- `sh scripts/test.sh` on the final tree, after the rebase and the review fixes: `test.sh: PASS, 7650 cases in 271s`.

## Size

- Estimated L, actual L.
