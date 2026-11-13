# Outcome: field-default-alloc

## Built

- `Expr.StructLit` in `selfhost/src/effects.nra` now walks each field default a literal leaves out (`walkLeftOutFieldDefaults`, `walkStructDefaults`, `findFieldDefaultKey`). The default is walked in its own `default|...` body and each effect it has (`alloc` for an array, map or string-building default) is charged at the literal with the cause `the default of <field>`, as `walkFilledDefaults` does for parameter defaults. A literal with a spread is not walked for defaults: the spread base supplies the other fields.
- The body key follows the emitter's `defaultBodyKey` (per-instance key for a generic field that uses a type parameter, `default|<module>.<name>|<field>` otherwise).
- Tests: `testdata/field_default_alloc.nra` (+ `.out`; `return {}` and a literal that gives every field, both with `uses alloc`), `testdata/errors/field_default_alloc.{nra,err}` (missing `alloc`).

## Proof

- base: feat/batch-13 (6ae5adcf), release compiler built from `git archive`; `narrac check testdata/errors/field_default_alloc.nra` exits 0 (no error for `export function make(): Bag { return {} }` with `items: Array<int> = [1, 2]`).
- branch: the same command reports `8:10: Function `make` uses `alloc` (through `the default of items`) ...`, exit 1. `narrac run -sanitize testdata/field_default_alloc.nra` prints `2`, `bag`, `0`, exit 0.

## Deviations

- None. Only `effects.nra`, the new test files and this node folder changed. `walkFilledDefaults` is not reused as is (it is keyed on call defaults); the new function has the same shape.

## Stubs left

- None.

## Gotchas

- A struct whose default is a constant that does not allocate (`int`, `string` literal) adds no effect, so existing code is unaffected.
- The key logic in `findFieldDefaultKey` duplicates `defaultBodyKey` in `emitexprs.nra`; change both together.

## Proposed learnings

- none

## Verification

- `narrac emit` of the base sources with the base and the branch compilers: byte-identical.
- `sh scripts/test.sh`: see the report line.

## Size

- Estimated S, actual S.
