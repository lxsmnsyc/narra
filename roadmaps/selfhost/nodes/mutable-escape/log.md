# Log: mutable-escape

- Checked first: no existing rule. Closure capture of a mutable parameter was already rejected (E0603).
- Rule: in the Go checker `noMutableEscape` (check_expr.go) rejects a bare identifier that is a `mutable` parameter when it is the value of `return`, a struct/array/tuple/map literal element, or the right side of an assignment whose target is a field or element. The selfhost checker has `checkNoMutableEscape` (exprs.nra), called from the same places in stmts.nra.
- Copies (`const q = p`, `x = p` to a plain variable, `p.field`, `f(p)`) stay allowed, as SPEC line 667 says.
- Code E0915, next free after E0914.
