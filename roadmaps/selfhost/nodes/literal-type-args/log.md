# Log: literal-type-args

- Nothing existed before (the brief's verify command passed only on old tests).
- Lookahead rule: after a capitalized `Ident`, a `<` is type arguments only if the tokens up to the matching `>` are names, `.`, `,`, `#[`, `]` and nested `<>` (`>>` closes two levels), and the next token is `{`. Chosen over trying the type parser and backing off, because Go's parser panics on error and the Narra parser has no rewind; a pure token scan is the same in both and has no side effects. Function types in the arguments are not accepted (`(` is not in the set); this keeps `A < (b)` out of the scan.
- Only an `Ident` that starts with an uppercase letter is considered, so `f<int>(x)` is not a form (it never existed) and `a < b` is untouched.
- Both `Map<..> {` and `Name<..> {` go through the existing `{` branch of `postfix`; the `<` case shares it.
- AST: `MapLit.TypeArgs` and `StructLit.TypeArgs` (`[]*TypeExpr`, shared by `clone` like other type expressions). Narra: `MapLit(id, pos, typeArgs, keys, vals)`, `StructLit(id, pos, typeName, typeArgs, fields, spread)`. Dump: `.typeargs` slot first, only when non-empty.
- Checker: the written arguments are resolved through `resolveType` with a synthetic `TypeExpr`, so arity, "not generic" and the float-key check reuse the annotation code. With a spread, `spreadLit` runs with the written type as the expected one and the result must equal it.
- New code E0237 for "type arguments disagree with the expected type".
- Dead end: first put the dump cases in `exprs.nra`; it shifted every line number in the golden. Moved them to the new `astdump/literal_args.nra`.
- Added `lit_targs_*` cases to `parseEdgeCases` in `selfhost_test.go` (outside the shared list; test data only).
