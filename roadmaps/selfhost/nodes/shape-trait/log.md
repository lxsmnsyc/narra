# Log: shape-trait

- The self-hosted compiler has one special case for `derives Shape`: `declareListed` in `selfhost/src/traits.nra` reports E0502 by the name `Shape` before looking the trait up. It stays (derive belongs to shape-derive) and does not clash with the source trait, since it runs on `derives` lists only.
- Implement blocks reject prelude and generic types (E0513, E0514), so the prelude impls are private generic Narra functions in `std/src/shape.nra`. `checkTraitCall` points a `Shape` call at them (`setShapeRef`) and makes the instance with `addFuncInstance`.
- First build failed: `typeString(checker, ...)` inside a `mutable checker` call; read into a local.
- A failing bound (E0530) still made an instance, which re-reported the error inside std/json. `checkBounds` now returns a bool and the call skips the instance.
- `lookupTrait(checker, "Shape")` is -1 in a module that did not import the trait; the trait index is passed in instead.
