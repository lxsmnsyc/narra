# Outcome: split-compiler

## Built

All in `compiler/internal/narra/`. Code was moved, not rewritten; function names are unchanged.

Checker files:
- `check.go`: `Program` (now with `Modules`), `scope`, `genericType`, `Checker`, `Check`, `errorf`, `run` (declaration passes, imports), scope helpers (`lookup`, `declare`, `push`, `pop`, narrowing).
- `check_types.go`: type resolution, generics (`instantiate`, `substitute`, `unify`), `orderTypes`, `arrayOf`, `expectType`, field and variant resolution, default values.
- `check_stmt.go`: `funcBody`, `block`, `ifStmt`, `stmt` (including the `for ... of` special cases for `Range.*` and `Array.values`), `lvalueRoot`, narrowing helpers.
- `check_expr.go`: `expr`, `exprInner`, `binary`, `structLit`.
- `check_call.go`: `call`, `genericCall`, `funcInstance`, `builtinCall`, `args`, mutable argument checks, enum constructors (`ctor`, `ctorArgs`).
- `check_match.go`: `match`, `resolvePattern`.

Emitter files:
- `emit.go`: `Emitter`, `Emit` (program layout), temporaries and scope unwinding, `function`, `signature`.
- `emit_types.go`: type definitions, `retain`/`release`, `rcFuncs`, `arrayFuncs`.
- `emit_stmt.go`: `stmt`, `forStmt`.
- `emit_expr.go`: `expr`, `binary`, `lvaluePtr`, literals.
- `emit_call.go`: `call`, `builtin` (looks up the registry), `argCode`.
- `emit_match.go`: `ifIs`, `bindPattern`, `matchArms`, `matchStmt`, `matchValue`.

Builtin registry (contract `builtins` v1, `builtins.go`): `Builtin` (with `Mutable`, `Effects`, `Check`, `Emit`), `Module`, `registerModule`, `registerPrelude`, `registerNamespace`. Lookup tables: `stdModules`, `preludeFuncs`, `namespaceFuncs`, `builtinsByName`. Each module registers in an `init()`:
- `std_int.go`: `std/int` (toString, abs).
- `std_float.go`: `std/float` (toString, sqrt, fromInt, truncate).
- `std_string.go`: `std/string` (length, concat, fromChar, repeat).
- `std_prelude.go`: `print`, `assert`, `panic`, `Array.length/values/push/filled`, `Range.upTo/between` and their checker functions.

To add a std function, add an entry to the module's `Funcs` map in its `std_*.go` file (`Params`, `Result`, `Emit`, or `Check` for generic ones). No edit to `check_call.go` or `emit_call.go` is needed. A new module is a new `std_<name>.go` with `registerModule`.

Per-module runtime C: `Module.Runtime` is written by `Emit` after `runtime.h` for each imported module (`Program.Modules`, once each, import order). Existing modules leave it empty because their helpers stay in `runtime.h`, so the emitted C is byte-identical.

Diagnostic (contract `diagnostic` v1, `diagnostic.go`): `Diagnostic`, `Fix`, `Edit`, `Diagnostics` (an `error` that prints one diagnostic per line), `(*Checker).report`. `errorf` builds `Diagnostic{Code: "E0000", ...}` and calls `report`. `Parse` wraps lexer and parser errors in one `Diagnostic` through `syntaxDiagnostic`. `Check` returns `Diagnostics`, so callers can type-assert to get structure.

Contracts: builtins v1 frozen, diagnostic v1 frozen (no change requests).

## Deviations

- `parser.go` (not owned) changed in `Parse` only, 8 lines, to wrap errors in a `Diagnostic`. Text is unchanged.
- `Diagnostic.Error` always prints the position, so errors without one still print `0:0:` as on main. `compiler/testdata/errors/contains_itself.nra` and `unknown_module.nra` pin the two messages.
- `registerModule` defaults `Builtin.Name` to `<module path>.<key>` when empty.
- `Diagnostic.File` is not filled: `driver.go` (not owned) still prefixes the file name with `prefix`. `json-diagnostics` should fill it there.

## Stubs left

- `Module.Runtime` is empty for all modules. Nodes that add runtime C (strings-std, io-std, maps) put it in their module.
- `Builtin.Mutable` is honored by the generic path in `builtinCall` but no builtin uses it yet; `Array.push` has its own `Check`.

## Gotchas

- `CallExpr.Builtin` stays a string, the builtin's `Name`. The emitter finds the entry in `builtinsByName`, so every `Name` must be unique (registration panics on a duplicate).
- `Range.*` and `Array.values` have no `Emit`; they are only valid in `for ... of`, which is special-cased in `check_stmt.go` and `emit_stmt.go`.
- Registration runs in `init()`; Go runs file `init`s in file name order, so do not depend on order between modules.
- Hot files for later nodes: `check_expr.go`, `emit_expr.go` (shared switches).

## Proposed learnings

- Add std functions only through `registerModule` in a `std_*.go` file, never in `check_call.go` or `emit_call.go` | keeps feature nodes from colliding | compiler/internal/narra | no
- Report compile errors through `c.errorf`/`c.report`, never by building strings | keeps `Diagnostic` the single path | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- A temporary test dumped the generated C (normal and test mode) for every `.nra` under `compiler/testdata/` before and after; all 48 outputs, including error texts, are identical. The test was removed.

## Size

- Estimated M, actual M.
