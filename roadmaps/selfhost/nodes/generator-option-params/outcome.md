# Outcome: generator-option-params

## Built

- `selfhost/src/generatorpairs.nra`: `collectGeneratorPairs(checker, want, got, mutable wants, mutable gots)` returns true when two types differ only in the effects of the generators inside them (through `Option` and other enum instances, structs, tuples, `Array`, `Box`, `Async`) and lists each pair of `Generator` or `AsyncGenerator` types.
- `selfhost/src/exprs.nra`: `expectType` calls it before the final mismatch error and checks each pair with `checkGeneratorResult` (immediate, or deferred to the effect pass for bound and inferred generators). `sameGeneratorShape` for function results uses it too, so function types with a nested generator result fit as well.
- A bound `Option<Generator<T> uses (E, alloc)>` now passes to a parameter, and a plain function can return it or hold it in a result type that writes the effects. `AsyncGenerator` works the same.
- When the effects do not fit, the error names the generators with their effects: `expected Generator<Array<int>> uses io, found Generator<Array<int>> uses alloc` (testdata/errors/generator_effect_types).
- Tests: `testdata/generator_effect_types.nra` (+ `.out`), `testdata/errors/generator_effect_types.nra` (+ `.err`).
- `testdata/iterator.nra`: `showGroups` takes the `Option` that `chunk` and `window` return, so the five call sites no longer unwrap it where it is made.

## Proof

- base: d64eee66, `narrac check testdata/generator_effect_types.nra` gives 8 E0119 errors (`expected Option<Generator<Array<int>> uses (alloc, io)>, found Option<Generator<Array<int>>>`); the error sample gives `expected Option<Generator<Array<int>> uses io>, found Option<Generator<Array<int>>>`.
- branch: the program sample checks and prints the expected output, and the error sample gives the effects error above.

## Deviations

- Case 2 of the brief (`expected Generator<int>, found Generator<int> uses alloc`) did not reproduce for a bare generator: `Generator<int> uses alloc from source` already worked on base, and `Generator<int> from source uses alloc` is rightly an error (the `uses` there is the function's). The nested form (`Option`, `Array`, tuple of generators) is what failed and is fixed and tested.
- A mismatch found before the effect pass, such as `Option<Generator<Array<string>>>` against `Option<Generator<Array<int>>>`, still prints the bound generator without effects, because a bound type has no effects until the effect pass fills them. Effects errors print them.
- std/iterator.nra needed no change; the simplification is in `testdata/iterator.nra`.

## Stubs left

- None.

## Gotchas

- A `TFunc` inside a nested type is not paired; it still compares exactly.
- The emitted C does not change: `narrac emit selfhost` from the base and the branch compiler is byte-identical.

## Proposed learnings

- Compare composite types with `collectGeneratorPairs` when generator effects may differ, not with `==` | `==` treats each effect list as another type | selfhost/src | no

## Verification

- `cmp` of `narrac emit selfhost` from the base and the branch compiler: identical.
- `sh scripts/test.sh --filter generator_effect_types` and `--filter iterator`: PASS. The full check is in the commit report.

## Size

- Estimated M, actual S. The check is one walk and one dispatch row.
