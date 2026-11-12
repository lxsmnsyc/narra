# Outcome: generic-implement

## Built

- Constraints on struct and enum type parameters: `struct Labeled<T: Show>`, `enum Tree<T: Ordered>`, `struct Grid<T: Show, const N: int>`. The parser keeps them in `StructDecl.bounds` and `EnumDecl.bounds` (`parseTypeDeclParams` in `selfhost/src/parser.nra`), astdump prints `.bound` slots, fmt and `narrac query` signatures print them. `declareModuleBounds` (`selfhost/src/typebounds.nra`) fills `Checker.bounds` for the type's parameters after declareTraits. Each written use (a type in `resolveType`, a struct literal, an enum constructor) queues a `TypeBoundCheck` through `noteTypeBounds` (`check.nra`); `flushTypeBounds` reports E1020 once every module is declared, with the hint `declare it as <T: Show>` for a type parameter argument. Bounds are not implied.
- Implement blocks on generic types: `implement Show for Pair<T> { ... }`. `checkImplementHead` (`selfhost/src/genericimpl.nra`) reports E1021 for `Pair<int>`, `Pair`, a wrong count or another name. The block sits in `checker.implementations` with `forType` the open type `Pair<T>` and in the type's `blocks`, so every instance implements the trait. Each member is also a generic `FuncInfo` named `<block index>_implement_<member>` with `selfType: Some(open type)`. A Narra name cannot start with a digit, so the C name cannot repeat a user's or another block's. Its body is checked once open (with the type's bounds and `Self`), then per instance, and emitted per instance.
- `findImplementation(mutable checker: Checker, type: Type, traitIndex: int, name: string): Option<ProgramFunc>` in `selfhost/src/implementations.nra`. It returns `Some(ProgramFunc.ImplFunc(block, index))` for a block on a concrete type, `Some(ProgramFunc.InstanceFunc(instance))` for a member of a block on a generic type (the instance is made on demand; `checker.instanceList[instance].funcIndex` is the member), and `None` when the compiler provides the function (automatic Equatable or Hashable, derives, Shape of prelude types) or the type's arguments fail its bounds. `lookupImplementation(checker, type, traitIndex, name)` gives the same answer without making instances, for the emitter. `setImplementationRef(checker, nodeId, program)` turns the answer into the call's ref (`Ref.Implementation`, or `Ref.Function` plus `instanceRefs`).
- `addImplementInstances` (called by `checkPending` in body.nra at the start and whenever the queue drains) gives every concrete instance of a generic type with blocks all member instances, so `writeDefinition` (emittraits.nra) calls the user's functions for `==`, map keys, derived `Ordered` and `Show` of enclosing types.
- Samples: `testdata/type_constraints.nra`, `testdata/type_constraints_joined.nra`, `testdata/generic_implement.nra`, `testdata/generic_implement_std.nra`, `testdata/generic_implement_names.nra` (types `A_B` and `A` with traits `C` and `B_C`, which an earlier name scheme merged). Errors: `testdata/errors/type_constraint_{written,literal,constructor,parameter}.nra`, `testdata/errors/generic_implement_{concrete,bare,count,name,missing,signature}.nra`. Goldens: `testdata/astdump/generic_implement.*`, `testdata/fmt/decl_generic_implement.*`, `testdata/parse/struct_bound.ast` (now a passing parse), `testdata/declsdump/traits/err_impl_target.dump` (E1021).
- E0021 and E0514 are gone from `diagnostics.nra` and the source; E1020 and E1021 have rows and explanations. SPEC "Generics" and "Traits" state the rules, the two README lines are removed, D106 is in DECISIONS.md with a "Built" line.

## Deviations

- `selfhost/src/check.nra` (not in owned or shared): `FuncInfo.selfType: Option<Type> = None`, the `TypeBoundCheck` struct, `Checker.typeBounds` and `noteTypeBounds`. resolve.nra cannot import traits.nra (traits imports resolve), so the use-site check is queued like `checkMapKey`; and an instance of a member needs `Self` in its environment. All additions have defaults, so no other constructor changed.
- Two new modules besides `typebounds.nra`: `selfhost/src/genericimpl.nra` (head check and member registration; imported by traits.nra, so it imports no traits) and `selfhost/src/implementations.nra` (findImplementation and the instance flush; imports traits, typebounds and instances). One module would form an import cycle.
- A type alias with bounds (`type P<T: Show> = ...`) used to hit E0021 too; it is now the parser's ordinary E0011 `expected ">", found ":"` (no new code). Recorded in D106.
- `unused.nra` (shared) skips members (`selfType is Some`) in the E0947 check.
- `selfhost/src/lspnavigation.nra` (not shared): `describeFunction` names a member `<module>.<Type>.<Trait>.<member>` through `describeImplementMember` (genericimpl.nra) instead of its hidden name.
- A rejected implement head (E1021, and E0117 on a non-generic type) marks the trait's entry in `TypeTraits.blocks` with -1, so E0507 does not ask for a block that was written. This applies to the base E0117 case too.
- A struct literal or enum constructor whose type arguments come from the expected type queues no E1020 check: the written type was already checked (no double report).

## Stubs left

- None.

## Gotchas

- A member's FuncInfo keeps the original `FuncDecl` (empty `typeParameters`, name `label`); only `FuncInfo.name` is hidden. Find members with `findImplementMember(checker, scope, pos)` or by `info.selfType is Some(_)`.
- Generic blocks are not in `checker.implFunctions` or `programFunctions`; anything that reads those for user implementations must go through `findImplementation` or `lookupImplementation`.
- `lookupImplementation` only finds instances the checker made. A new emitter-side use of a trait on a generic instance must be backed by a checker-side instance; `addImplementInstances` covers every concrete instance type in `checker.structs` and `checker.enums`.
- `TypeTraits.blocks[trait]` can be -1 for a rejected block. `is Some(_)` readers treat it as implemented, which keeps cascades out; a reader that indexes `checker.implementations` must skip negative values (findGenericMember does).
- Any new place that prints `FuncInfo.name` must map members through `describeImplementMember` or use `declaration.name`; E0310 now uses `declaration.name`.

## Proposed learnings

- A check in resolve.nra that needs trait facts is queued on the checker and flushed beside flushMapKeys | resolve.nra cannot import traits.nra, and trait facts of later types are not known yet | selfhost/src | no

## Verification

- `NARRAC=<dev narrac> sh scripts/test.sh --update`, then `--suite fmt --update` (new goldens).
- Every new sample runs in the programs suite in debug (sanitizers, leak check) and release.
- `sh scripts/test.sh` after the rebase and review fixes: `test.sh: PASS, 7292 cases in 134s`.

## Size

- Estimated L, actual L.
