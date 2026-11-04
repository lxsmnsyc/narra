# Log: selfhost-emit-closures

- Ported emit_closure.go into selfhost/src/emitclosures.nra in one pass.
- Function types come from Checker.composites (TFunc), sorted by typeId with a local stable insertion sort (emitdecls' mergeSort is private).
- Arrow bodies stay as written in the AST; an expression body is wrapped in a ReturnStmt or ExprStmt at emit time, as Go's checker does.
- Arrow parameters are found as `info.locals[info.declared[param.id - info.first]]`.
- The "// Closures" section is written only when the prototype buffer is non-empty (Go: closures != nil).
- Narra rejects `writeLine(mutable emitter, ... localRef(emitter, ...))`: read into a local first.
- Compared whole outputs against Go: effects.nra and param_after_uses.nra are byte-identical. closures.nra and closures_nested.nra still differ only through std and match stubs (temp and literal numbers shift).
