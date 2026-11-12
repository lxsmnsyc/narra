---
id: generic-implement
title: "implement blocks and constraints on generic types"
status: active
priority: should
type: feature
story: batch-11
early: false
outline: false
milestone:
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/generic-implement/**
  - selfhost/src/typebounds.nra
  - testdata/generic_implement*
  - testdata/type_constraints*
  - testdata/errors/generic_implement*
  - testdata/errors/type_constraint*
shared:
  - selfhost/src/emittraits.nra
  - selfhost/src/body.nra
  - selfhost/src/query.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/querysignature.nra
  - selfhost/src/declare.nra
  - selfhost/src/generics.nra
  - selfhost/src/resolve.nra
  - selfhost/src/exprs.nra
  - selfhost/src/ctors.nra
  - selfhost/src/traits.nra
  - selfhost/src/calls.nra
  - selfhost/src/instances.nra
  - selfhost/src/unused.nra
  - selfhost/src/diagnostics.nra
  - testdata/parse/struct_bound.*
  - testdata/declsdump/traits/err_impl_target.*
  - testdata/astdump/**
  - testdata/fmt/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E1020, diag E1021, decision D106]
provides: []
depends_on: []
---

# implement blocks and constraints on generic types

## Goal

Two README differences go away. Type parameters of structs and enums take trait constraints (`struct Labeled<T: Show>`), and a generic struct or enum gets `implement` blocks (`implement Show for Pair<T> { ... }`), as SPEC section "Traits" allows for any type.

Errors on the current compiler (built from the seed at `57d55087`):

- `struct Labeled<T: Show> { value: T }` and `enum Tree<T: Ordered> { Leaf  Node(value: T) }`: E0021 `trait constraints on struct and enum type parameters are not supported by this compiler yet` (parser, `parseTypeParamsNoBounds`).
- `struct Pair<T> implements Label { left: T  right: T }` with `implement Label for Pair<T> { function label(value: Pair<T>): string { ... } }`: E0514 `implement blocks for generic types are not supported by this compiler yet; use derives` (`resolveImpl`), then E0507 `Pair implements Label but has no implement Label for Pair block`, then E0524 `Pair<int> does not implement Label` at `Label.label(Pair { left: 1, right: 2 })`. Through a bound (`describe<T: Label>(Pair { ... })`) it is E0525. The same for a generic enum.

Plan, in two steps on one branch:

1. **Constraints on type parameters.** The parser keeps the bounds of struct and enum type parameters in `StructDecl` and `EnumDecl` (a `bounds` field like `FuncDecl.bounds`). The declaration pass puts them in `checker.bounds` for the type's `GenericInfo.parameters`, so field types and implement bodies see `T: Show`. Each written use of the type with arguments (`Labeled<int>` in a type, a struct literal, an enum constructor, an inferred instance) checks the arguments against the bounds: a failure is E1020. A type argument that is itself a type parameter needs the bound declared on its own parameter (`function first<T>(pair: Pair<T>)` with `struct Pair<T: Show>` is E1020, with the hint `declare it as <T: Show>`). Bounds are not implied.
2. **implement on generic types.** `implement Trait for Pair<T>` names the type with its own parameter names, in order. The members are checked once with `T` open (with the type's bounds), then per instance, like generic functions. A call through the trait on `Pair<int>` (direct or through a bound) resolves to the member's instance.

## Context

- SPEC: "Traits" (around line 749), "Generics" (line 712). D5 (implementations live with the type). README lines to remove: "Constraints are allowed on the type parameters of functions only" and "`implement` blocks work for non-generic structs and enums only".
- Parser: `parseTypeParamsNoBounds` in `selfhost/src/parser.nra` (E0021). `parseStructDecl` and `parseEnumDecl` call it; `parseTypeParams` already returns `bounds`. AST: `StructDecl` and `EnumDecl` in `selfhost/src/ast.nra` gain `bounds: Array<Array<TypeExpr>> = []`. This is a new field, not a new node kind, but `astdump.nra` (the `bound` slot, as for functions), `fmt.nra` (`printTypeParameters` already prints bounds; struct and enum pass `[]` today) and `querysignature.nra` (struct and enum signatures print names only) must print it. `testdata/parse/struct_bound.*` holds today's E0021 golden: it turns into a passing parse.
- Declarations: `declare.nra` makes the type's parameters with `makeParameters` (around line 321); `funcBounds` in `traits.nra` is the model for filling `checker.bounds`. `typeImplements` on a `TParam` then reads the bounds with no change.
- Constraint checks: `instantiate` in `resolve.nra` has no position and is also called by `substitute` for instance bodies, where the check is redundant. Check at the sites that have a position: `resolveType` (the two `instantiate` calls near lines 668 and 766), struct literals (`exprs.nra:2017`), enum constructors (`ctors.nra:280`). Put the shared check in a new module `selfhost/src/typebounds.nra` (`checkTypeBounds(checker, pos, generic, arguments)`), modeled on `checkBounds` in `traits.nra`, which reports E0525 for functions.
- implement on generic types: `resolveImpl` in `traits.nra` (E0514, around line 655) and `checkListed`, which passes `Type.TVoid` for a generic type today. `nominalImplements` keys type facts by `<module>|<name>`, so a block recorded on the generic entry already makes every instance implement the trait; the type's bounds hold for every instance by step 1.
- Calls: `checkTraitCall` in `calls.nra` finds the member with `implFunc` (`traits.nra`), keyed by `typeKey(type)`, so `Pair<int>` finds nothing. Recommended path: register each member of a generic implement block as a generic function in `checker.functions`, with the block's parameter names and the type's bounds, and resolve a trait call on an instance to `Ref.Function` plus `addFuncInstance(funcIndex, instanceArguments)` (`instances.nra`). This is how `setShapeRef` reaches the `derives Shape` functions of generic types, so body checks, effects, emission and `narrac query` work unchanged. Factor the lookup into one exported helper (for example `findImplementation(checker, type, traitIndex, name)`) that both `checkTraitCall` and later the Dynamic tables (node dynamic-traits) call.
- `unused.nra` (E0947) must not report the registered members as unused private functions.
- `testdata/declsdump/traits/err_impl_target.nra` has `implement Show for Gen {}` on a generic type: its golden changes from E0514 to the new head error.

## Steps

- [ ] Constraints: parser, AST field, astdump, fmt, querysignature, declaration of bounds, `checkTypeBounds` at the use sites (E1020).
- [ ] Generic implement: head check (E1021), member registration and instances, trait calls on instances, direct and through bounds.
- [ ] SPEC, README, D106, goldens.

## Acceptance

- [ ] Samples (`testdata/type_constraints*`): a constrained struct and enum used with `int`, `string` and a user struct; a field of type `T` passed to `Show.show` inside a function over `Labeled<T>` whose own `T` is bounded; two traits joined with `&`; a constrained type with a constant parameter (`struct Grid<T: Show, const N: int>`).
- [ ] Samples (`testdata/generic_implement*`): `implement` for a generic struct and a generic enum; a user trait and `Show`; a call on a concrete instance and through `<T: Label>`; a member that calls a trait function on a field of type `T` (needs the type's bound); an instance nested in another (`Pair<Pair<int>>`); the same type with an `implement` block for one trait and `derives` for another.
- [ ] Every sample is clean under the leak check and sanitizers.
- [ ] Error tests: a type argument that fails a bound in a written type, a struct literal and an enum constructor (E1020); a generic function using `Pair<T>` without the bound (E1020 with the hint); `implement Show for Pair<int>`, `implement Show for Pair` (no arguments), the wrong argument count and a name other than the declaration's (E1021); a missing member (E0521) and a wrong signature (E0520) on a generic block.
- [ ] E0021 and E0514 are gone from the code table in `diagnostics.nra` and from the source. The `codes` suite passes.
- [ ] astdump and fmt goldens for bounds on structs and enums and for a generic implement head. `narrac fmt` is stable on them.
- [ ] SPEC.md "Traits" and "Generics" state the rules; the two README lines are removed; D106 is recorded with a "Built" line.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

L. Scores not low: files touched (parser to emitter path, about 15 shared files), two features in one node, golden churn in astdump, fmt and declsdump. If it runs long, land step 1 alone and split step 2 into a follow-up node.

## Notes for the worker

- Open design questions, with the recommended answer (record them in D106):
  - **Head syntax.** `implement Show for Pair<T>` writes the declaration's own parameter names, in order. A concrete type (`Pair<int>`), a missing list, the wrong count or another name is E1021. Recommended because it keeps one name per parameter across the declaration and its blocks and rules out specialization. Rejected: Rust's `impl<T>` binder (new syntax) and fresh names mapped by position (two names for one parameter).
  - **Where bounds come from.** Only the type declaration. An implement block adds no bounds of its own, so there are no conditional implementations (`Show` for `Pair<T>` only when `T: Show`). A type that wants `Show` for every instance declares `struct Pair<T: Show>`. Conditional implementations can come later on top of this.
  - **Implied bounds.** None: `function first<T>(pair: Pair<T>)` must write `<T: Show>`. Matches the explicit style of the rest of the language and keeps `typeImplements` on a `TParam` a plain lookup.
- Functions inside an implement block still take no type parameters (E0517 stays).
- `derives` on a constrained type keeps working: the derived functions see the bounds through the parameters.
- The trait lookup helper is what node dynamic-traits builds on. Keep its name and signature in your `outcome.md`.
- No deprecation paths: E0021 and E0514 are removed, not kept as hints.

## Plan review (2026-11-12)

- The emitter does not work unchanged. `writeDefinition` (emittraits.nra:249-251) finds a user block through `implFunctions[typeKey(needType)|Trait.fn]`. Once members of a generic block live in `checker.functions`, `useShow`, `useEq`, `useHash` and `useCmp` on `Pair<int>` would fall back to the derived function (`==`, map keys, `Show.show` through `<T: Show>` in `emitTraitCall` emittraits:146-190, sorting). The checker must create the member instance wherever a std trait is used on such an instance, and `writeDefinition` must call the new `findImplementation`.
- Code that reads `checker.implementations` by block changes too: `body.nra:672-690` (implementation results), `query.nra:814` (hover names) and `declsdump.nra:278-282`. Either keep generic blocks in `checker.implementations` as well, or update these readers.
- Write `findImplementation(checker, type, traitIndex, name)` and its result type into your outcome, because dynamic-traits builds its function tables on it.
- distinct-traits starts after this node lands (it changes the same `useTraitFunction` and `writeDefinition` path, emittraits:90-108).
