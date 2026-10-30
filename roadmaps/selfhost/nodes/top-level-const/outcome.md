# Outcome: top-level-const

## Built

- `compiler/internal/narra/check_const.go`: `declareConsts`, `importConst`, `constType` (lazy check, cycle detection), `moduleConstValue` (`Alias.NAME`), `constWrite`, `checkConstValue`.
- `compiler/internal/narra/emit_const.go`: scalar and fieldless-constructor constants are emitted at each use. Others get `G_<hash>_<name>` (storage, plus a ready flag) and a getter `g_<hash>_<name>()` that builds the value on first use and returns it borrowed. `narra_consts_release` frees them before `narra_report_leaks`.
- `ConstDecl` (ast.go), `Ident.Const`, `FieldExpr.Const`, `Program.Consts`, `p.constDecl`, `modScope.consts`. `export const`, imports by name (`import { N }`) and `Alias.N` work. A constant holding a function can be called, also through `Alias.f(...)`.
- Allowed values: literals, operators, template strings, constructors, struct/array/map literals, spreads, function names, other constants.
- Tests: `testdata/const.nra`, `testdata/packages/const` (constants imported by name and by alias, across two modules), errors `const_call`, `const_cycle`, `const_top_assign`, `const_mutable_arg`, `const_twice`, `const_match`.

Diagnostic codes taken: E0190 (declared twice), E0191 (cycle), E0192 (calls a function), E0193 (other non-constant value), E0290 (change a constant).

## Deviations

- Section 10 says no initialization code. Non-scalar constants are built on first use by a getter, not at startup and not as static C data. Recorded in `compiler/README.md`. The first-use build is not thread safe.
- Package test is in `testdata/packages/const` (the package harness reads that folder); the single-file test is `testdata/const.nra`.
- Files outside owns and shared: `modules.go` (exports and `declName` cases), `check_expr.go`, `check_call.go`, `check_stmt.go` (a few lines each), `emit_expr.go` (two cases), `cnames.go` (`constCName`, `constGetCName`), `diagnostic_codes.go`. `clone.go` needed no change.

## Stubs left

- `@const` functions and `lazy` are not supported. Struct and array constants are not static C data.
- Operators on constants (`1 / 0`, overflow) panic at first use, not at compile time.

## Gotchas

- Resolution order for a name: local, constant, function. A constant and a function cannot share a name.
- Do not reset `c.closure` when checking a constant: function types registered there would be lost.

## Proposed learnings

- Emit a constant read only through `e.constRef` | the getter returns a borrowed value | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (samples under ASan, UBSan, leak check, release).

## Size

- Estimated M, actual M.
