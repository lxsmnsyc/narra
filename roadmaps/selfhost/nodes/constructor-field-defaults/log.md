# Log: constructor-field-defaults

- Branch fix/constructor-field-defaults, reset to feat/batch-13 (0bab063f).
- Proof on base: `narrac types testdata/constructor_field_defaults.nra` with a release compiler built from 0bab063f reports `constructor_field_defaults.nra:8:21: a field default must be a constant` for `kind: Kind = Kind.None`. `Kind.Sized(4)` and `Some(Kind.Sized(-1))` already passed: a call sets its Ref.Constructor on the call node.
- Cause: a bare qualified constructor is an `Expr.FieldExpr`, and `isConstantValue` only accepted `Expr.Ident`. The checker records Ref.Constructor on the field node, so the fix adds FieldExpr to the same arm.
- Error test: `Kind.Sized(count())` is E0106 on both base and branch (a call to a function is not constant).
- Emit check: `narrac emit selfhost` from the base and the branch compilers on the same source is byte-identical.
