# Outcome: std-set

## Built

- `std/set` (`std/src/set.nra`, registry row in `selfhost/src/std/stdset.nra`, one `registerAll` line in `stdlib.nra`, alias `Sets`). Display signatures are in `docs/reference/std/set.md`:
  - `create<T: Hashable>(): Set<T> uses alloc`
  - `add<T: Hashable>(mutable set: Set<T>, item: T): bool uses alloc` (true when the item was new)
  - `has<T: Hashable>(set: Set<T>, item: T): bool`
  - `remove<T: Hashable>(mutable set: Set<T>, item: T): bool uses alloc` (true when the item was there)
  - `size<T: Hashable>(set: Set<T>): int`
  - `values<T: Hashable>(set: Set<T>): Array<T> uses alloc` (insertion order)
  - `union`, `intersection`, `difference<T: Hashable>(first: Set<T>, second: Set<T>): Set<T> uses alloc`: new sets, arguments unchanged.
- Design choices:
  - `Set<T>` is an `opaque struct` over `Map<T, bool>`, like `std/arena`. Outside the module the map cannot be read or written (`testdata/errors/set_members_private`), so no code can store `false` and break the "every value is true" rule, and the representation can change later.
  - It has value semantics through the field: `copy(set)` is independent (a case checks it).
  - `T` is bound by `Hashable`, the trait map keys need.
  - `values` returns an array, as `Maps.keys` does, so it works in `for ... of`. Result order follows the map: insertion order.
  - `union`, `intersection` and `difference` build a new set through `add`; `intersection` and `difference` keep the order of the first argument.
- Review nits from std-array-more: the helper `smaller` in `std/src/ext__array.nra` is now `pickSmaller`; `testdata/errors/array_contains_not_equatable` (a function-typed element rejected by `contains`).
- Tests: `testdata/std_set.nra` (empty set; add new and again; remove and remove missing; copy independence; union, intersection and difference both ways, with an empty set, and arguments unchanged; string and struct items), errors `set_members_private` and `array_contains_not_equatable`.
- Guide: chapter 13 has a new section "Distinct values: `std/set`" with a runnable sample; chapter 8's std module list and the chapter 13 line of `docs/guide/README.md` name `std/set`. The std reference page `docs/reference/std/set.md` is generated (`scripts/std/generate.nra` module table has the row); `selfhost/src/std/stdtext.nra` is regenerated. The unknown-module message lists `std/set` (from the registry).

## Deviations

- None.

## Stubs left

- Hand-written `Map<_, bool>` sets replaced: none. `testdata/` and `scripts/` have no map used as a set (`scripts/unicode/generate.nra` `excluded` and `replaced`, and `scripts/std/generate.nra` `done`, are arrays of bool indexed by position, not sets). Only `Map<_, bool>` literals that test the map itself remain (`testdata/astdump/literal_args.nra`, `testdata/typesdump/patterns/literals.nra`).
- Left for after the seed is regenerated (`selfhost/src` and `selfhost/tests/src` cannot use `std/set` yet): 42 files declare `Map<_, bool>`: `selfhost/src/` `tools/lspcache`, `tools/fmt`, `driver/lockfile`, `driver/manifest`, `driver/depresolve`, `driver/driver`, `driver/hash`, `driver/modules`, `std/stdcopy`, `emit/emitexprs`, `emit/emitdecls`, `emit/emitsimd`, `emit/emit`, `emit/emitclosures`, `emit/cnames`, `emit/emitmaps`, `emit/emitstate`, `check/constfunctions`, `check/resolve`, `check/match`, `check/traits`, `check/sharedresults`, `check/effects`, `check/unused`, `check/declare`, `check/ctors`, `check/scopes`, `check/sharedwrites`, `check/asynccheck`, `check/exprs`, `check/types`, `check/typebounds`, `check/generics`, `check/threadcheck`, `check/body`, `check/check`, `check/unsafecode`, `syntax/literals`, `syntax/trivia`, `syntax/ast`, `syntax/parser`, and `selfhost/tests/src/emitchecks`. Some of these map to values other than a set; check each before replacing.

## Gotchas

- `Sets.values` copies the keys into an array on each call, so `has` in a loop is cheaper than scanning `values`.
- Removing an item and adding it again moves it to the end of the order, as a map does.
- A new std source module needs the registry file, the `registerAll` line, the module row in `scripts/std/generate.nra`, then `--suite generated --update`; the compiler built before that step has no `std/set`.

## Proposed learnings

- A collection type in `std/src` is an opaque struct over its representation | keeps invariants and allows a later change | std/src | yes (an error case reads a field)

## Verification

- `sh scripts/test.sh --suite generated --update` (std reference and `stdtext.nra`), `narrac run testdata/std_set.nra`, `narrac check` on both new error cases.
- `--update` of the errors and dumps suites: the module list in the unknown-module goldens gained `std/set`, and the typesdump goldens follow the `pickSmaller` rename.
- Full `sh scripts/test.sh`: `PASS, 10883 cases in 325s`. `pnpm --dir website build`: passes.

## Size

- Estimated S, actual S.
