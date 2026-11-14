# Log: field-shorthand

- Proof, base 6c616a46 (feat/batch-15), compiler built from the seed with `-release`: `narrac run testdata/field_shorthand.nra` prints `field_shorthand.nra:35:18: write `x: x`; field shorthand is not allowed` and exits 1.
- Design: the parser turns `name` into `FieldInit { isShorthand: true, value: Expr.Ident at the field's position }`, so the checker, effects, emitter and const evaluator need no change. The unknown-name hint lives in `selfhost/src/fieldshorthand.nra` (two call sites in `exprs.nra`).
- E0037 is kept as a retired row (as E0022). No code of E1280-E1284 was needed.
- The language server only publishes diagnostics and `narrac query` has no field or rename sites, so the hover, go-to-definition and rename items have nothing to attach to. The value is a name at the field's position, so `type-of` on it reaches the local.
