# Outcome: shape-trait

## Built

- `std/src/shape.nra`: `export trait Shape { toValue(value: Self): Value; fromValue(value: Value): Result<Self, string> }` and `implement Shape for Value`. `Value` lists `implements Shape`. The prelude implementations are private generic functions in the same file (`intToValue`, `arrayFromValue<T>`, `mapToValue<V>`, `optionFromValue<T>`, `tuple2ToValue<A, B>` to `tuple4...`).
- `std/src/json.nra`: `encode<T: Shape>(value: T): string` and `decode<T: Shape>(text: string): Result<T, string>`; call `Json.decode<User>(text)`.
- Compiler: `shapeImplements`, `shapeProblem`, `reportShapeProblem`, `isShapeTrait` in `traits.nra`; `setShapeRef` and `findShapeProvider` in `calls.nra` (a `Shape` trait call on a prelude type becomes `Ref.Function` of the std/shape function, with an instance from `addFuncInstance`). `checkBounds` now returns a bool and a call whose bounds fail makes no instance (`calls.nra`, `funcvalues.nra`), so the generic body is not re-checked with the bad type.
- Mechanism for shape-derive: generate or write ordinary Narra `toValue`/`fromValue` functions and point the trait call at them, as `setShapeRef` does; no C is emitted for Shape.
- Samples: `testdata/shape.nra/.out` (hand-written `implement Shape` for a struct and a nested struct, every prelude type, every decode error form, ASan and leak check clean); `testdata/errors/shape_map_key`, `shape_function`, `shape_struct`, `shape_call` (`.nra` and `.err`).
- Docs: SPEC.md (traits table, "Serialization with `Shape`"), compiler/README.md, DECISIONS.md D84.

## The `derives Shape` special case

The self-hosted compiler reports E0502 for `derives Shape` in `declareListed` (`traits.nra`), by name, before any trait lookup. It does not clash with the source trait and stays until shape-derive replaces it.

## New codes

- E0530 `<type> does not implement Shape: <reason>` (with explanation). It replaces E0524 (direct `Shape.toValue(x)` call) and E0525 (a bound `T: Shape`) when the trait is Shape. The reason names the failing part: a Map key that is not `string`, a function, a struct or enum without an `implement` block, a tuple of other than two to four items.

## Behavior to know

- Error path format: `.users[2].age: expected int, found Text`; at the root `expected int, found Text`. Containers add `.key` or `[index]` in front of the child message; a child message that starts with `.` or `[` is taken as a path.
- `f64` accepts `Int` and `Float`; `int` only `Int`; `char` a Text of one character; `Option<T>` decodes `Null` as `None`; extra record keys are ignored.
- Sized integers (`u8` ...) do not implement Shape; tuples above four items do not. Both are E0530.
- `Option<Option<T>>` is ambiguous (`Null` decodes as the outer `None`).

## Gate decisions

- Prelude impls as private Narra generic functions in std/shape, found by the checker: alternatives were compiler-generated C (more code) and lifting E0513/E0514 for implement blocks on generic and prelude types (bigger change).
- E0530 replaces E0524/E0525 for Shape (alternative: keep those codes and use E0530 only for Map keys and functions); chosen so one code names the failing part.
- Tuples limited to two to four items; sized integers left out (not in the brief).
- Errors carry paths as text prefixes, not a structured path type.

## Stubs left

- None. `derives Shape` is shape-derive.

## Size

- Estimated M, actual M.
