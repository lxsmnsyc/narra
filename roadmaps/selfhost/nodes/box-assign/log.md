# Log: box-assign

- Checker already rejects `b.value = x` on a `const` box ("is constant"); only the emitter was broken.
- Emitter: `lvaluePtr` of a FieldExpr on a Box now calls `<box>_at(&box)`, which runs `<box>_unique` (copy with retained payload when rc > 1) and returns `&f_value`. Same in `selfhost/src/emitexprs.nra`; helpers added to `boxFuncs` / `boxFunctions`.
- Selfhost gotcha: `"emitter" passed as mutable` forbids `containerStruct(emitter, ...)` inside the same call as `mutable emitter`; hoist it into a const.
