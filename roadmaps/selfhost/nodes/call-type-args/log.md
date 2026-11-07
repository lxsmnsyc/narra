# Log: call-type-args

- Go: `CallExpr.TypeArgs`; parser `callTypeArgsAhead` (balanced `<...>` of type names, closing `>` directly before `(`).
- Ambiguity rule: a name inside the arguments that is not followed by `.` must start with a capital letter or be a primitive type name. This keeps `a < b > (c)` a comparison, which TypeScript does not.
- Checker: `call` wraps `callTarget`; `genericCall` and `traitCall` take the type arguments (`takeTypeArgs`) and mark them used. A call with type arguments that nobody took reports E0326.
- Trait declarations now accept `Self` only in the result (the old E0512 check counts the result).
- Selfhost: `Expr.CallExpr` gained `typeArguments` after `callee`; every pattern got a `_`. `Walker.checkCall` is unchanged: the type arguments travel in `Checker.callTypeArguments` keyed by call node id, emptied by whoever takes them, so the signature stays out of the way of `effects-indirect-calls`.
- Query site index: type arguments are not indexed, as for literal type arguments.
