# Outcome: distinct-traits

## Built

- `distinct type UserId = int implements Label derives Ordered, Show`: the parser (`parseOpaqueOrDistinct`) reads the lists; `DistinctDecl.implementations` and `.derived` are printed by astdump, `fmt` (`printDistinct`) and `writeDistinct` in `querysignature.nra`.
- `traits.nra`: `declareTraits` makes a `TypeTraits` entry for each distinct type; `typeTraitsKey` handles `TDistinct`; `distinctImplements` returns true for an own `implement` block, a derived trait whose base implements it, and the automatic Equatable and Hashable. `checkListed` reports E1012 when the base does not implement a derived trait (Shape too). `shapeDeriveHolds` follows the base.
- `emittraits.nra`: `stripDistinct` replaces `underlyingType` in `useTraitFunction` and stops at a distinct type with its own block, so a generic instance with `T = UserId` calls the user's `Show`.
- `derives Shape`: `deriveshape.nra` writes `__shapeToValue_<Type>` and `__shapeFromValue_<Type>` through the base's Shape; `setShapeRef` in `calls.nra` points at them. The unwrap inside uses the hidden name `__unwrap` (accepted by `checkDistinctCall`).
- E1012 row and explanation in `diagnostics.nra`. SPEC "Distinct types", README (two lines removed, one reworded), D108, D62 "Built" note.
- Samples: `testdata/distinct_traits.nra`, `testdata/packages/distinct_traits_modules`. Errors: `testdata/errors/distinct_traits_{base,missing_block,both,user_derive,less}`, `testdata/packages/errors/distinct_traits_other_module` (E0513). Goldens: `testdata/astdump/distinct_traits`, `testdata/fmt/decl_distinct_traits`, `testdata/query/basic` (UserId now lists traits).

## Deviations

- `testdata/packages/distinct_traits_modules` and `testdata/packages/errors/distinct_traits_other_module`: a package is the only way to test another module, and `testdata/distinct_traits*` does not cover that folder.
- `testdata/query/basic/**` changed for the query signature golden.
- `calls.nra` accepts the hidden unwrap name `__unwrap` on a distinct type only inside the generated `__shapeToValue_*` functions (`isInShapeToValue`); a user call is E0322 (test `distinct_traits_hidden_unwrap`).
- E1014 (reserved by the coordinator, not in the brief): `implement Equatable` or `implement Hashable` for a distinct type is rejected, since `==` and map keys use the base value. Tests `distinct_traits_equatable` and `distinct_traits_hashable`.
- `E0508` test is "derives and implements the same trait" (the existing meaning), not a trait in both lists of E0501/E0504.

## Stubs left

- None.

## Gotchas

- A derived Show on a distinct type prints the base text, so a distinct `string` is quoted.
- On a struct, a user `implement Equatable` is honored by `==` (checked); on a distinct type it is E1014.

## Inbox

- `__Shape`, `__Value` and `__ShapeText` (hidden imports of generated Shape code, `deriveshape.nra`) may be reachable from user code by name; not checked or fixed here.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: `test.sh: PASS, 7572 cases in 268s`.

## Size

- Estimated M, actual M.
