# Outcome: std-array-more

## Built

- `std/array` gains seven functions. Display signatures are in the std reference (`docs/reference/std/array.md`) and `narrac query interface std/array`:
  - `truncate<T>(mutable items: Array<T>, length: int)`: a builtin (row in `selfhost/src/std/stdarray.nra`, check `checkArrayTruncate`, emitted helper `<array>_truncate`, evaluator case in `check/consteval.nra`). It keeps the first `length` items and releases the rest. A `length` of at least the current length does nothing and never copies; a negative `length` is 0. A shared array is copied once, as `push` does (value semantics).
  - `pop<T>(mutable items: Array<T>): Option<T> uses alloc`: Narra in `std/src/ext__array.nra`. It copies out the last item (a reference count bump, not an array copy) and calls `truncate`. `None` for an empty array.
  - `indexOf<T: Equatable>(items, wanted): Option<int>`, `contains<T: Equatable>(items, wanted): bool`, `reverse<T>(mutable items)`: Narra.
  - `sort<T: Ordered>(mutable items)` and `sortBy<T, effect E>(mutable items, compare: (T, T) => Ordering uses E)`: Narra, a stable bottom-up merge sort with one buffer of the array's length.
- Design choices:
  - Sorting is in place (`mutable items`), like `push`, `pop` and `reverse`: the call site shows which array changes, a shared array is copied only when it is shared, and a caller who wants to keep the old order writes `copy(items)` first. A returning form would need a second allocation for every call.
  - `sortBy` takes a comparison function returning `Ordering`, not a key function: it covers descending order and several keys, and `Int.compare` and `Ordered.compare` already have that shape. `sort` calls `sortBy` with `Ordered.compare`.
  - `pop` is not a runtime builtin of its own: `truncate` is the only new primitive, since `pop` is `truncate` plus one item copy.
- Tests: `testdata/std_array_more.nra` (sort of ints, strings, a derived `Ordered` struct, empty and one-item arrays; `sortBy` stability with equal groups and a descending order; copy independence after `sort`, `pop` and `truncate`; `pop` on ints, strings and an empty array; `truncate` past the end and negative; `indexOf`, `contains`, `reverse` on odd, even and empty arrays). Error cases `testdata/errors/array_truncate_not_mutable`, `array_truncate_not_array` and `array_sort_not_ordered`. `query/builtin_signatures` probes the `truncate` signature; `testdata/query/std_modules` shows the new functions.
- Guide: chapter 2 has a new section "Sort, search and shrink" with one runnable sample; chapter 11's folder sample prints the sorted names (the inbox entry about sorting is answered). The std reference is regenerated.
- Diagnostics: E0906 now reads `%s needs an array, found %s` (it said `Array.push`), and `truncate` reports E0905 and E0906 with its own name.
- Hand-written copies replaced: `scripts/unicode/generate.nra` `sortPairs` and `sortMultiple` (insertion sorts, now `Arrays.sortBy`), `testdata/traits.nra` `contains`.

## Deviations

- `pop` is Narra over `truncate`, not a second builtin (see above).
- `selfhost/tests/src/files.nra` `sortNames` and `main.nra` `contains` stay: the test runner is built by the seed compiler, so it cannot call the new functions yet.

## Stubs left

- Compiler call sites to replace after the seed is regenerated: `driver/modules.nra` (`sortStrings`, `popLast`), `driver/textops.nra` (`dropLast`), `driver/workspace.nra` `sortNames`, `tools/querygraph.nra` `sortStrings`, `std/builtins.nra` `sortStrings`, `check/ctors.nra` `sortNames`, `syntax/triviadump.nra` `sortKeys`, `check/resolve.nra` `sortEffects`, `check/effects.nra` `sortEffectNames`, the `mergeSort` copies in `emit/emit.nra`, `emit/emitdecls.nra` and `emit/emitmaps.nra`, `tools/audit.nra`, `driver/driver.nra` `sortDiagnostics`; the `scopeUndoCount` counter beside an array in the checker; `std/arena`'s `freeCount` stack; `selfhost/tests/src/files.nra` `sortNames`; `selfhost/tests/src/main.nra` `contains`.

## Gotchas

- `sort` and `sortBy` copy each item (a reference count bump for heap items) when they move it, so a sort of large strings pays a retain and release per move.
- `Arrays.pop(mutable items)` cannot sit in a template string that also reads `items`; write the result to a `const` first.
- The seed has no `truncate` and `selfhost/src` cannot use any of these functions until the seed is regenerated.

## Proposed learnings

- A std function that moves array items in place copies each with `copy(...)` and writes through `items[index]` | the move keeps counts right for heap items and works on shared arrays | std/src | no

## Verification

- `sh scripts/test.sh --suite generated --update`, `--suite query`, `--suite guide` (612 cases), `--suite errors`: pass.
- Full `sh scripts/test.sh`: `PASS, 10842 cases in 231s`. `pnpm --dir website build`: passes.

## Size

- Estimated M, actual M.
