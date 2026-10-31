# Log: selfhost-check-patterns

- Work already existed? No: `is`, match, destructuring, Map/tuple/spread and Map index were all `unsupported:` stubs.
- Layout: patterns.nra (below exprs: narrowing, irrefutable, exhaustiveness), match.nra and literals.nra (above exprs, reached from exprs through new Walker trait functions). Needed because resolving a pattern calls expectType from exprs.
- Narrowing is a `narrow|<local>` key in the block scope stack (`Checker.localScopes`), so pushScope/popScope save and restore it. FieldExpr records `Ref.Narrow`, patterns record `Ref.Variant`.
- `isBinds` is not a Checker field: checkIf calls checkIs with binds=true when the whole condition is an `is`.
- Go dump: pattern lines get `ctor=` and a `local` line after a binding pat-name.
- First run of TestSelfhostPatterns: no diffs on the new inputs or the earlier ones.
