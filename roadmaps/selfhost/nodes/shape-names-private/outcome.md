# Outcome: shape-names-private

## Built

- `selfhost/src/deriveshape.nra`: `addShapeDerives` returns `Result<Parsed, Diagnostic>`. In a file that derives Shape it lexes the source and reports E0203 `unknown name "<name>"` at the first identifier that is `__Value` or starts with `__Shape` or `__shape` (covers `__Shape`, `__ShapeText`, `__shapeToValue_<Type>`, `__shapeFromValue_<Type>`, `__shapePath`, `__shapeKind`), and for a template literal that contains one.
- `testdata/errors/shape_hidden_{value,text,shape,to_value,from_value,template}`: error tests.

## Proof

- base: af35fdcb, `narrac check probe.nra` with `struct P derives Shape { x: int }` and a `run` that uses `__Value.Null`, `__ShapeText.length("a")`, `__shapeToValue_P(...)`, `__shapeFromValue_P(...)`, or `f<T: __Shape>`. Result: each type-checks, with only an unused-variable warning.
- branch: the same probes report `unknown name "__Value"` and so on at the name. `NARRAC=<built> sh scripts/test.sh --suite errors --filter shape_hidden`: PASS, 6 cases.

## Deviations

- `selfhost/src/modules.nra` (not owned): turns the new Err of `addShapeDerives` into a `LoadError`, as for a parse error.
- The error is E0203 from the loader, not a checker rule like `__unwrap`'s E0322: a checker rule would need hooks at value, type and bound lookups.

## Stubs left

- None.

## Gotchas

- A name inside a plain string or comment is allowed. A name inside an interpolated template is rejected, reported at the template start under the matching prefix.
- Only files that derive Shape are scanned; elsewhere the names do not exist and the usual unknown-name errors apply.

## Proposed learnings

- Hidden generated names are rejected by scanning the source of the file that receives them | the generated code shares the file's scope | selfhost/src/deriveshape.nra | yes

## Verification

- `sh selfhost/bootstrap.sh`: fixed point reached (seed older than source, as expected).
- `sh scripts/test.sh`: see the final line in the report.

## Size

- Estimated S, actual S.
