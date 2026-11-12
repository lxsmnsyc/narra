---
id: distinct-traits
title: "implements and derives on distinct types"
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
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-12
spent:
owns:
  - roadmaps/selfhost/nodes/distinct-traits/**
  - testdata/distinct_traits*
  - testdata/errors/distinct_traits*
shared:
  - selfhost/src/emittraits.nra
  - selfhost/src/ast.nra
  - selfhost/src/parser.nra
  - selfhost/src/astdump.nra
  - selfhost/src/fmt.nra
  - selfhost/src/trivia.nra
  - selfhost/src/querysignature.nra
  - selfhost/src/traits.nra
  - selfhost/src/deriveshape.nra
  - selfhost/src/calls.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/diagnostics.nra
  - testdata/astdump/**
  - testdata/fmt/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E1012, decision D108]
provides: []
depends_on:
  - on: generic-implement
    kind: hard
    why: both change useTraitFunction and writeDefinition in emittraits.nra (emittraits:90-108) and trait lookup in traits.nra and calls.nra; underlyingType strips a distinct type before the key is built, so this node changes the path generic-implement rewrites
---

# implements and derives on distinct types

## Goal

A distinct type lists traits like a struct or an enum: `distinct type UserId = int implements Label derives Ordered, Show`. A listed trait needs an `implement` block in the same module or a `derives` entry; a derived trait uses the base type's implementation, since the C type is the same. `==`, `!=` and map keys keep working through the automatic `Equatable` and `Hashable` (D62), and SPEC is corrected to say so.

Errors on the current compiler (built from the seed at `57d55087`):

- `distinct type UserId = int derives Show`: `dist2.nra:1:28: expected the end of the statement, found "derives"` (E0013, `parseOpaqueOrDistinct` in `parser.nra`). The same for `implements`.
- Without a list, `implement Show for UserId { ... }` is E0513 `cannot implement Show for UserId: the type is declared in another module; wrap it in a struct`, a wrong message: the type is in the same module, but `declareTraits` makes no `TypeTraits` entry for a distinct type.
- `Show.show(UserId.from(1))` is E0524 `UserId does not implement Show`.
- `UserId.from(1) == UserId.from(1)` compiles and runs today; `UserId.from(1) < UserId.from(2)` is E0215.

Plan: the parser reads the trait lists after the base type, `declareTraits` makes a `TypeTraits` entry for each distinct type, `typeImplements` answers for a `TDistinct` from that entry, derived traits check the base and emit the base's functions, and `implement` blocks work as for structs.

## Context

- SPEC section 4 "Distinct types" (line 631) and "Traits" (line 803, which already allows `implements` on a distinct type). SPEC line 656 "No operators carry over" contradicts line 714 (`Equatable` for every type without a function inside) and D62. README "Differences from the spec" lines 224 and 225 go away. D62's cost line ("No `implements` or `derives` on distinct types yet") gets a "Built" note.
- Parser: `parseOpaqueOrDistinct` (`parser.nra:1467`) calls `parseType` for the base, then `endTypeDecl`. Call `parseTraitLists` (`parser.nra:2043`) between them. `DistinctDecl` (`ast.nra:260`) gains `implementations: Array<TypeExpr>` and `derived: Array<TypeExpr>`, as `StructDecl` has. This is a new field, not a new node kind, but these must print it: `astdump.nra:397` (the same slots structs use), `fmt.nra:974` `printDistinct` (the lists after the base, as `printStruct` writes them), `trivia.nra` if the keywords need trivia keys (see `selfhost/CLAUDE.md` on comments around keywords), and `querysignature.nra:194` `writeDistinct`.
- Checker: `declareTraits` (`traits.nra:157`) adds entries for `Decl.Struct` and `Decl.Enum`; add `Decl.Distinct` with `generic: false`, and call `declareListed` for it. `typeImplements` (`traits.nra:932`) returns true for a `TDistinct` only for std `Equatable` and `Hashable` through the base; send the other traits to `nominalImplements` (`traits.nra:1091`), and make `findTypeTraits` find the distinct's entry. In `nominalImplements`, a derived trait on a distinct type holds when the base implements it (the base takes the place of the field walk).
- A `derives` entry whose base does not implement the trait is E1012 at the entry, the distinct type's version of E0509 (which names a field). The derivable traits stay the E0503 list: `Equatable`, `Hashable`, `Ordered`, `Show`, `Shape`.
- Shape: `shapeImplements` (`traits.nra`) and `deriveshape.nra` handle `derives Shape`; `setShapeRef` in `calls.nra` resolves `Shape.toValue` and `Shape.fromValue`. A derived `Shape` on a distinct type converts through the base's `Shape` (a `UserId` encodes as `Value.Int`).
- Emitter: derived traits need no new C. `useTraitFunction` (`emittraits.nra:90`) strips the type with `underlyingType` (`emitstate.nra:430`), so a derived `Show` or `Ordered` on `UserId` already calls the `int` function. A user `implement` block is the trap: a trait call written directly resolves to `Ref.Implementation` and is fine, but a generic instance `<T: Show>` with `T = UserId` reaches `useShow`, which strips the distinct type and would call the base's `Show`. Stop the strip for a trait that the distinct type implements with its own block.
- Related nodes in this batch: generic-implement and dynamic-traits also edit `traits.nra` and `calls.nra` (hot files, see GOAL.md). dynamic-traits builds `Dynamic.of` tables through the same trait lookup, so a distinct type that implements a trait should work there with no extra case.

## Acceptance

- [ ] Samples (`testdata/distinct_traits*`): `derives Show` and `derives Ordered` on a distinct `int` and a distinct `string` (`Show.show`, `Ordered.compare`, sorting an `Array<UserId>` through a generic `<T: Ordered>` function); `implements Label` with an `implement` block that reads the base through `UserId.toInt`; a user `implement Show for UserId` called directly and through `<T: Show>` (the user text, not the base's); `derives Shape` with `Json.encode` and `Json.decode<UserId>`; `==` and a `Map<UserId, string>` key; an exported distinct type with traits used from another module; `export opaque distinct type` with `derives Show`.
- [ ] Every sample is clean under the leak check and sanitizers.
- [ ] Error tests (`testdata/errors/distinct_traits*`): `derives Show` on a distinct type over a base without `Show` (a `Ptr<int>` from `std/ptr`, or a struct without `Show`) is E1012; `implements Label` with no block (E0507); a trait listed in both lists (E0508); `derives Label` for a user trait (E0503); `implement Show for` a distinct type of another module (E0513, with its message unchanged); `UserId.from(1) < UserId.from(2)` stays E0215.
- [ ] E1012 is a row in `diagnostics.nra`; the `codes` suite passes.
- [ ] astdump, fmt and query signature goldens for a distinct type with both lists. `narrac fmt` is stable on them.
- [ ] SPEC.md "Distinct types" states the trait lists, what a derived trait does, and that `==` and map keys come from the automatic `Equatable` and `Hashable` while arithmetic and `<` do not carry over. D108 records the open questions below with a "Built" line; D62 notes that its cost line is built. README lines 224 and 225 are removed.
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: one parser change, one AST field pair printed in three dumps, a trait-table entry and a fix to the emitter's strip of distinct types. No new type or node kind.

## Notes for the worker

- Open design questions, with the recommended answer (record them in D108):
  - **Where the lists go.** After the base: `distinct type UserId = int implements Label derives Show`. It reads left to right like a struct head (`struct Money implements Label {`), the base stays next to `=` as in a type alias, and `implements` is a keyword, so a generic base such as `Array<int>` ends without ambiguity. Rejected: `distinct type UserId implements Label = int`, which splits the name from its base.
  - **`==` on a distinct type.** Keep it (D62). SPEC line 714 already gives every type without a function inside an automatic `Equatable`, and a distinct type is such a type. Fix line 656 to say that arithmetic, bit operators and `<` do not carry over. Rejected: following line 656 literally, which would take away `==` and map keys from every distinct type and make ids such as `UserId` hard to use.
  - **What derived `Show` prints.** The base's text: `Show.show(UserId.from(7))` is `7`, and a distinct `string` is quoted as `Show.show` quotes a string. A distinct type has the base's representation and no fields, so there is no `UserId(7)` form to show. Rejected: `UserId(7)`, which reads like an enum constructor that does not exist.
  - **`derives Shape` on an opaque distinct type.** Allowed: the declaring module chose to derive it, as with an opaque struct. Note in SPEC that `Shape.fromValue` then builds values outside the module.
- E1012 message: `cannot derive %s for %s: its base type %s does not implement %s`.
- A user `implement Equatable` block on a distinct type behaves as the same block on a struct does today; do not change how `==` picks its function.
- No deprecation paths.

## Plan review (2026-11-12)

- Starts from generic-implement. Use its `findImplementation` so a user `implement Show for UserId` block is found before `underlyingType` replaces the type with its base, including through `<T: Show>`.
