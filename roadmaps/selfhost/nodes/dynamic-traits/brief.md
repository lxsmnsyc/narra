---
id: dynamic-traits
title: "Dynamic<Trait> values"
status: todo
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
added: 2026-11-04
spent:
owns:
  - roadmaps/selfhost/nodes/dynamic-traits/**
  - selfhost/src/dynamic.nra
  - selfhost/src/emitdynamic.nra
  - testdata/dynamic_traits*
  - testdata/errors/dynamic_trait*
shared:
  - selfhost/src/types.nra
  - selfhost/src/check.nra
  - selfhost/src/resolve.nra
  - selfhost/src/generics.nra
  - selfhost/src/calls.nra
  - selfhost/src/traits.nra
  - selfhost/src/cnames.nra
  - selfhost/src/emit.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/sharedresults.nra
  - selfhost/src/stdcopy.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E1022, diag E1023, diag E1024, diag E1025]
provides: []
depends_on:
  - on: generic-implement
    kind: hard
    why: both rewrite trait-call resolution in calls.nra, traits.nra and emittraits.nra (hot files), and the Dynamic tables must call the implementation lookup that generic-implement extends to generic instances; built in parallel they conflict and a Dynamic of a generic type would miss its implementation
---

# Dynamic<Trait> values

## Goal

D95 "Dynamic traits", SPEC "Trait values" (lines 815-817). `Dynamic<Trait>` holds a value of any type that implements one trait. `Dynamic.of(value)` makes one when the expected type is known, and `Drawable.draw(item, mutable canvas)` calls the implementation of the stored value's own type. So `Array<Dynamic<Drawable>>` holds circles and squares.

Errors on the current compiler (built from the seed at `57d55087`), with `trait Drawable { function draw(value: Self): string }` and two implementing structs:

- `const items: Array<Dynamic<Drawable>> = [Dynamic.of(Circle { radius: 1 }), Dynamic.of(Square { side: 2 })]`: E0117 `type Dynamic is not generic` at the type, E0308 `unknown module "Dynamic"; import it first` at each `Dynamic.of`, then E0522 `cannot tell which type's Drawable.draw to call here` at `Drawable.draw(item)` in the loop.
- `function render(item: Dynamic<Drawable>): string { return Drawable.draw(item) }`: E0117, then E0522.
- The parse error the spec-gap audit saw (`expected a name, found "of"`) is gone: `of` is accepted after `.` now, so no parser change is needed.

`Hash.of` does not belong here: it is built (`std/src/hash.nra`, `Hash.of(7)` and `Hash.of(Point { ... })` run). The README "Not supported yet" line `Dynamic<Trait> and Hash.of` is stale for `Hash.of`; this node removes the whole line.

Plan:

- **Type.** A new `Type.TDynamic(traitIndex: int)` (`types.nra`), resolved from `Dynamic<Trait>` in `resolveType` (`resolve.nra`, next to `Box`), with `typeKey`, `typeString` (`Dynamic<Drawable>`), `unify` and `substitute` cases. `Dynamic` is a prelude name, like `Box`: a user type named `Dynamic` hides it (`userType`).
- **Which traits.** Per D95: every function of the trait takes `Self` as its first parameter, exactly once and not nested (`Array<Self>` does not count), and no function mentions `Self` anywhere else, the result included. So `Show` and user traits like `Drawable` qualify; `Equatable`, `Ordered`, `Hashable` (Self twice, or used by `==` and map keys) and `Shape` (`fromValue` returns `Self`) do not. Other trait functions may take `mutable` parameters of other types (`mutable canvas: Canvas`).
- **`Dynamic.of(value)`.** Checked in `calls.nra` where the `unknown module` error is reported today, when `Dynamic` is not a user name. The expected type must be a `Dynamic<Trait>`; the argument's type must implement the trait (E0524 as for a trait call). In a generic body `<T: Drawable>`, `Dynamic.of(value)` with `value: T` works and is emitted per instance.
- **Calls.** In `checkTraitCall`, a first argument of type `Dynamic<Trait>` for the same trait binds `Self` to the Dynamic. The call keeps `Ref.TraitFunction` with `traitSelf` set to the `TDynamic`, so the `Ref` switches in query, effects and dumps need no new case. Its effects are the trait function's declared effects, as for a call through a bound.
- **Representation.** One allocation per `Dynamic.of`: a reference-counted cell that holds a pointer to the table and the value. A `Dynamic` is a pointer to that cell, so a copy is one retain and the value is never copied. The table, one per (trait, concrete type), is a static const struct of C function pointers: a `drop` that releases the value and frees the cell, and one thunk per trait function that reads the value out of the cell and calls the implementation (a user block, a `generic-implement` instance, or the compiler's own `Show` for primitives and derived types through `useShow`). `Drawable.draw(item, ...)` emits `item->table->draw(item, ...)`.

## Context

- Decisions: D95 (design), D81 rule 8 (a `Dynamic` cannot be passed to a `mutable Self` function: E1025 at the call), D5, D12 (no hidden calls: the call through the table is written as a trait call). SPEC line 1231 repeats the D81 rule. README "Differences from the spec" has a line saying that rule has nothing to check yet: remove it, keeping the identity-type half.
- Checker: `checkTraitCall` and `setShapeRef` in `selfhost/src/calls.nra`, `typeImplements`/`computeImplements` in `traits.nra` (a `TDynamic` implements no trait for bounds; see the open questions), `isContentsType` in `check.nra` (a `Dynamic` is not a contents type, so `mutable d: Dynamic<X>` is already rejected by rule 1). Put the checks in a new module `selfhost/src/dynamic.nra`: the trait rule (E1023, reported at the `Dynamic<Trait>` type and naming the trait function that breaks it), `Dynamic.of` (E1024), the `mutable Self` check (E1025).
- Implementation lookup: use the helper generic-implement exports (see its `outcome.md`) to find each table entry, so a `Dynamic.of(Pair { ... })` of a generic type with an `implement` block works.
- Emitter: `cnames.nra` (C name of the cell type, like `Box`), `emitstate.nra` and `emitdecls.nra` (retain, release and the slot functions, near the `TBox` cases), `emittraits.nra` `emitTraitCall` (the table call), `emit.nra` (where tables are written: after prototypes, before bodies). Put the table, thunk and `Dynamic.of` emission in `selfhost/src/emitdynamic.nra`.
- `Type` switches without a `_` arm need a `TDynamic` case: `sharedwrites.nra`, `sharedresults.nra`, `stdcopy.nra` (a `copy(d)` retains, the value stays shared and read-only), `declsdump.nra`, `typesdump.nra`. The compiler reports the missing ones.
- `Show.show` on a `Dynamic<Show>` is a plain trait call through the table.

## Acceptance

- [ ] Samples (`testdata/dynamic_traits*`): an `Array<Dynamic<Drawable>>` of two struct types looped with `for ... of`; a `Dynamic` as a parameter, a result, a struct field and inside `Option`; a trait function with extra parameters, one of them `mutable` of another type (SPEC's `Drawable.draw(item, mutable canvas)`); `Dynamic<Show>` of an `int`, a `string` and a derived struct; a `Dynamic` of a generic struct with an `implement` block (generic-implement); `Dynamic.of(value)` in a generic body `<T: Drawable>`; a copied `Dynamic` used after the original; a trait imported from another module in a package sample.
- [ ] Every sample is clean under the leak check and sanitizers, including a `Dynamic` dropped while copies remain and an array of them released at the end.
- [ ] Error tests: `Dynamic<int>`, `Dynamic` with no argument and with two (E1022); `Dynamic<Ordered>`, `Dynamic<Shape>` and a user trait with `Self` in a result or not first (E1023, naming the function); `const item = Dynamic.of(circle)` with no expected type (E1024); `Dynamic.of(3)` into a `Dynamic<Drawable>` (E0524); a `Dynamic` passed to a `mutable Self` trait function (E1025); `==` on two `Dynamic` values (the existing equality error) and a `Dynamic` as a `Map` key (the existing key error).
- [ ] SPEC.md "Trait values" states the trait rule, the one-allocation layout and the D81 rule; README drops `Dynamic<Trait> and Hash.of` from "Not supported yet" and the D81 half-line from "Differences from the spec". D95 gets a "Built" line for dynamic traits with the amendments.
- [ ] `narrac query` hover shows `Dynamic<Drawable>` for such a local.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

L. Scores not low: a new `Type` variant (many switch sites), a new runtime layout with reference counting and generated thunks (leak risk), checker and emitter in one node.

## Notes for the worker

- Open design questions, with the recommended answer (record them as D95 amendments, no new decision number):
  - **Does `Dynamic<Drawable>` satisfy `<T: Drawable>`?** No in this node: a bound sees a `Dynamic` as not implementing the trait (E0525 as today). Allowing it would let a bounded body pass a local of type `T` as `mutable Self` (D81 rule 8 allows that for bounds) with a `Dynamic` behind it. Revisit with a rule for that case.
  - **Does `Dynamic<Show>` implement `Show` so `Show.show(items)` works on an `Array<Dynamic<Show>>`?** No, for the same reason; loop and call `Show.show(item)`.
  - **`Dynamic` name.** A prelude type with one function, `Dynamic.of`, as SPEC writes it, not a `std/dynamic` module. `Dynamic.as<T>` (downcast) stays out (D95).
  - **Where the table pointer lives.** In the cell, not next to the pointer: copies stay one word and one retain. SPEC's "a box, plus a pointer to the implementation" reads either way; state the layout in SPEC.
- Calls through a `Dynamic` are not direct C calls, so the README line "every trait call is a direct call" needs "except calls through a `Dynamic`".
- The value in the cell is read-only: no field of it can be written through the `Dynamic`, which is why sharing it between copies is safe.
- Thunks must release what they own on every path; run the leak check on each sample while working (`NARRAC=<built narrac> sh scripts/test.sh --suite programs --filter dynamic`).

## Plan review (2026-11-12)

- Build the function tables on generic-implement's `findImplementation` (see its outcome for the result type). Its generic-instance sample (`Dynamic<Show>` over `Pair<int>`) is the last step.
