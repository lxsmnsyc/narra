# Outcome: opaque-distinct

## Built

All in `compiler/internal/narra/`; main logic in `check_opaque.go`.

- Syntax: `export opaque struct` (`StructDecl.Opaque`), `distinct type A = B` and `export opaque distinct type A = B` (`DistinctDecl`). `opaque`, `distinct` and `type` are contextual words (`Parser.isWord`). `.from` is accepted as a member name although `from` is a keyword (`Parser.memberName`).
- Types: `DistinctType{Name, Module, Base, Opaque}`. Its C type is the base type's (no runtime cost). Its identity in composite names is `typeID` = `declCName("D", module, name)`, so `Array<UserId>` and `Array<int>` differ. `underlying(t)` strips distinct wrappers; the emitter uses it for retain, release, trait functions and `==`.
- Opaque struct rules (outside the declaring module): literal (`E0320`), field read or assign (`E0235`), spread (`E0236`), struct pattern (`E0720`). `StructType.Opaque` is set on declaration and on generic instances.
- Distinct rules: `T.from(x)` and `T.toInt(x)` (named after the base: `toString`, `toU64`, ...), checked in `distinctCall`, emitted as the argument itself (`CallExpr.Distinct`). Also through a namespace (`Alias.T.from`). `from` on an opaque distinct type outside its module is `E0321`. Mixing two distinct types, or a distinct type and its base, uses the ordinary type mismatch error; `+` etc. use the ordinary "needs two numbers" error. `==`/`!=` and map keys work when the base supports them (`implements`, `containsFunc`, `mapKeyProblem`).
- Tests: package `testdata/packages/opaque_range` (opaque structs with int, string and array fields, spread inside the module, distinct int and string, an array of distinct values, `==`, opaque distinct), and error packages `testdata/packages/errors/{opaque_literal,opaque_read,opaque_assign,opaque_spread,opaque_pattern,distinct_mix,distinct_int,distinct_arith,opaque_distinct_from}`.
- `compiler/README.md`: Supported and Not supported lines.

Diagnostic codes taken: E0043, E0044, E0045 (parser), E0120 (declarations), E0235, E0236 (expressions), E0320, E0321, E0322 (calls), E0720 (patterns). Next free: E0046, E0121, E0237, E0323, E0721 (other nodes may have taken some; check the table).

## Deviations

- Files outside owns and shared: `check.go`, `check_expr.go`, `check_match.go`, `check_patterns.go`, `check_traits.go`, `modules.go`, `modules_check.go`, `emit_expr.go`, `emit_traits.go`, `cnames.go`, `std_map.go`, `diagnostic_codes.go`, `compiler/README.md`. Each is a few lines (a case, a check or a table row).
- `UserId.from` and `UserId.toInt` are checker and emitter special cases, not registered builtins (they depend on the user's type).
- Non-exported `opaque` is a parse error (E0043), since opacity only exists for exported types.

## Stubs left

- No `implements`/`derives` on distinct types; no `Show` or `Ordered` for them. No generic distinct types.
- `Shape` for opaque structs does not exist yet (no Shape support).
- Cross-module function calls that return a hidden struct still work (as the spec wants); only field-level access is rejected.

## Gotchas

- Any new `switch t.(type)` over `Type` that handles base kinds should call `underlying` first or add a `*DistinctType` case.
- A distinct type must be declared before another distinct type that is its base (E0120 otherwise).

## Proposed learnings

- Strip distinct types with `underlying` before switching on a type in the emitter | a distinct type shares the C type of its base | compiler/internal/narra | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.

## Size

- Estimated M, actual M.
