# Outcome: shape-derive

## Built

- `selfhost/src/deriveshape.nra` (new): `addShapeDerives(parsed, source)` writes Narra text for every struct and enum of a package file that lists `derives Shape`, parses it with ids after the file's and positions after its last line, and appends the declarations and two hidden imports (`import { Shape as __Shape, Value as __Value } from "std/shape"`, `import * as __ShapeText from "std/string"`). Per type: `__shapeToValue_<Type>` and `__shapeFromValue_<Type>` (generic types get `<T: __Shape>`); per file: helpers `__shapePath` and `__shapeKind`. Called from `loadFile` and `loadPackage` in `modules.nra`; std source modules are not touched.
- `calls.nra` `setShapeRef`: a Shape call on a struct or enum with no implement block points at the generated function of the type's module, with an `addFuncInstance` instance for a generic type.
- `traits.nra`: `derivedTrait` (a bare `derives Shape` finds std/shape's trait through the hidden import when the module has no `Shape`), E0503 now allows Shape, E0531 in `checkListed` (generic types: the fields that mention no type parameter), E0532 in `checkShapeTags`, E0533 in `checkShapeVisible` (called from `checkTraitCall` and `checkBounds`), `shapeDeriveHolds`, and `nominalShapeProblem` (the E0530 reason for a derived type with a bad field).
- `body.nra` `checkModuleBodies`: skips the generated functions of a type whose derive is in error, so the error is reported once.
- Samples: `testdata/derive_shape.nra/.out` (defaults, Option fields, nested structs, arrays and maps of derived types, unit and fielded constructors, a recursive enum, generic struct and enum, exact JSON, round trips, error paths), `testdata/packages/derive_shape_modules` (use from another module without importing Shape; opaque struct used in its own module), `testdata/errors/derive_shape_field`, `derive_shape_type_field`, `testdata/packages/errors/derive_shape_opaque`. All run clean under `-sanitize` (ASan and leak check).
- Docs: SPEC.md section 4 (traits table, "Serialization with `Shape`"), compiler/README.md (Supported; removed from Not supported).

## New codes

- E0531 `cannot derive Shape for %s: field %q has type %s, which does not implement Shape: %s` (at the `Shape` entry of `derives`, like E0509).
- E0532 `%s derives Shape, so %s.%s cannot have a field named "type": ...` (at the field).
- E0533 `%s is an opaque struct that derives Shape; its Shape is usable only in its own module` (at the call).
- E0502 is no longer reported; its row stays (codes are never reused).

## Behavior to know

- Error texts added: `.field: missing field`, `expected Record, found <Kind>`, `.type: missing field; expected a constructor of <Enum>`, `.type: expected Text, found <Kind>`, `.type: <Enum> has no constructor "<name>"`.
- Option-ness is read from the type as written (`Option<...>`); an alias of an Option is treated as a required field.
- Names starting with `__shape`, and the aliases `__Shape`, `__Value`, `__ShapeText`, are taken in a module that derives Shape.
- E0533 is not checked inside generic instance bodies; the call site that bound the type was checked.
- `derives Shape` inside a std source module is not supported (std modules are not passed through the generator).

## Gate decisions

- Generated Narra text parsed by the loader, versus building AST nodes in the checker (needs ids and side tables after parse) or emitting C (new emitter paths, as shape-trait avoided). Text reuses the checker and emitter unchanged.
- Free functions plus `setShapeRef`, versus a generated `implement Shape for T` block: blocks would clash with E0508 and generic implement blocks are E0513/E0514.
- Hidden aliased imports, versus exported helper functions in std/shape (new public API) or a new `std/internal/shape` module (internal imports are E0426 for user modules). The two small helpers are written once per file instead.
- A bare `derives Shape` works without importing Shape (alternative: require the import, E0505). The module still imports Shape to call `Shape.toValue` itself.
- E0531 lists every bad field (alternative: the first only); E0533 at each use site, also through `Array<Span>` and other containers and type arguments.

## Suites run

- dumps (goldens updated for E0502/E0503), errors, codes, programs, packages, sources, std-tests, bootstrap, query, lsp: all pass. No full `scripts/test.sh` run, per the orchestrator.

## Stubs left

- None.

## Size

- Estimated M, actual M.
