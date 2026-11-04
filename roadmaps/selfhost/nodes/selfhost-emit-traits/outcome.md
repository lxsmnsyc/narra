# Outcome: selfhost-emit-traits

## Built

- `selfhost/src/emittraits.nra` (about 400 lines): `useEq`, `useHash`, `useCmp`, `useShow` (primitives return the runtime name; other types join `emitter.traitNeeds`, deduplicated through `traitSeen`), `emitTraitFunctions` (the `// Trait functions` section, with Go's index loop over the growing needs list), `emitEquality` (`==` and `!=` on structs, enums, arrays, tuples), `emitTraitCall` (eq, hash, cmp, show). Definitions call a user implementation from `Checker.implFunctions` when one exists, else write the eq, hash, cmp or show body for box, array, struct (tuple form `#[...]`) and enum.
- The skip file `compiler/testdata/emit/skips/selfhost-emit-traits.txt` is deleted (it listed 17 inputs, none now reports this node).
- Tuple `==` in `patterns_tuples` and `patterns_tuple_shapes` now reaches the literal pool correctly. Whole narrac output is byte-identical to Go for `traits`, `traits_derive`, `patterns_tuples`, `patterns_tuple_shapes`, `recursive`, `closures_traits`, `traits_functypes`, `patterns_nested`, `patterns_generic`, `trait_method_generic_return`, `traits_ctor_shadow`.

## Deviations

- `selfhost/src/calls.nra` (checker, not owned): `checkTraitCall` now records `bodies[body].traitSelf[nodeId] = selfType` next to `setRef(... Ref.TraitFunction ...)`. The contract gives the field to emit-program but nothing filled it (same situation as `instanceRefs` in emit-bodies).
- The skip file lists were not shrunk but deleted, as the brief says.

## Stubs left

- None in this node. `hasWork`'s "selfhost-emit-traits" arm in `emitstate.nra` is now unused; the final `selfhost-emit` node can remove it.

## Gotchas

- Narra rejects a template literal nested inside `${...}` of another template (`${cString(`${x}`)}`): compute the inner text into a local.
- Enum field access in eq, hash, cmp and show keeps Go's odd `a.u->v_<variant>.f_<field>` form, so output matches byte for byte.
- `useTraitFunction` strips distinct types first (Go's `underlying`), and looks up user implementations by `typeKey` of the stripped type.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostEmit`: pass.
- `cd compiler && go vet ./... && go test -count=1 ./...`: pass.

## Size

- Estimated M, actual M (about 400 Narra lines).
