# Outcome: generator-function-type

## Built

- A `function*` can be stored in a binding, constant or parameter of an annotated function type, with the generator effects checked. Changes in `selfhost/src/exprs.nra`, `effects.nra`, `cnames.nra`, `emitclosures.nra`.
- Sample `testdata/generator_function_type.nra` with `.out`; error samples `testdata/errors/generator_function_type_element.nra` and `generator_function_type_effect.nra`.

## Proof

- base: 6478acd, `narrac run testdata/generator_function_type.nra`: three errors `expected (int) => Generator<int>, found (int) => Generator<int>` (top-level, local, argument).
- branch: same command prints the values; a wrong element type still errors (`Generator<string>` vs `Generator<int>`); a generator with `io` stored in a pure type errors.

## Deviations

- The `alloc` of creating a generator is not charged against the annotated function type when its result is a generator, because the type cannot list it apart from the generator's effects.
- Generator effects no longer appear in the C name of a type, so function types differing only in effects share one C struct.

## Stubs left

- None.

## Gotchas

- The effect error for a function type with a generator result prints the generator types (`expected Generator<int>, found Generator<int> uses io`), not the whole function types.
- A `function*` with a `for` over `Ranges.upTo` has no generator effect (the creation `alloc` is a function effect), so only the `io` case is an error sample.

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh`: PASS (all shards, bootstrap job included).

## Size

- Estimated S, actual S.
