# Log

- Branch fix/shape-names-private reset to feat/batch-12 (af35fdcb).
- Probes on feat/batch-12, narrac from `sh selfhost/bootstrap.sh`, `narrac check` of a file with `struct P derives Shape`: user code naming `__Value`, `__ShapeText`, `__shapeToValue_P`, `__shapeFromValue_P` and a bound `T: __Shape` all type-check (only unused-variable warnings). All reachable.
- Fix: `addShapeDerives` (deriveshape.nra) lexes the source of a file that derives Shape and reports E0203 `unknown name` at the first hidden identifier or template containing one. `__unwrap` used a checker rule; templates and signatures made a checker rule too diffuse. modules.nra gets the Err path.
