# Outcome: literal-type-args

## Built

- Syntax: `Map<K, V> {}`, `Map<K, V> { ... }`, `Pair<A, B> { ... }` and `Pair<A, B> { ...s, a: 1 }`.
- Go parser (`parser.go`): `typeArgsLitAhead` and `litTypeArgs`, called from the `{` branch of `postfix`, which now also handles `<`.
- Narra parser (`selfhost/src/parser.nra`): the same two functions. `mapLit` and `structLit` take the type arguments.
- AST: `MapLit.TypeArgs`, `StructLit.TypeArgs` (`ast.go`, `clone.go`); Narra `ast.nra`: `MapLit(id, pos, typeArgs, keys, vals)` and `StructLit(id, pos, typeName, typeArgs, fields, spread)`. `idcheck.nra` patterns updated.
- Dumps (`astdump.go`, `astdump.nra`): a `.typeargs` slot with type nodes, printed first and only when non-empty. Doc comment updated. New golden `testdata/astdump/literal_args.{nra,ast}`.
- Checker: `mapLit` (std_map.go) and `structLit` (check_expr.go) resolve the arguments with `resolveType`, so arity, "not generic" and float-key errors reuse annotation code. A mismatch with the expected type is E0237 (`the type arguments make %s, but the expected type is %s`).
- Lookahead rule: after a capitalized name, `<` starts type arguments only if the tokens to the matching `>` are names, `.`, `,`, `#[`, `]` and nested `<>` (`>>` closes two levels), and the next token is `{`. Otherwise it is less-than.
- Tests: `testdata/literal_type_args.{nra,out}` (run, debug and release, under ASan and the leak check); eight error tests `testdata/errors/literal_type_args_*.nra` (map arity, struct arity, float key, value mismatch, field mismatch, map and struct disagree with the expected type, not generic); six `lit_targs_*` cases in `parseEdgeCases`; `literal_type_args_test.go` asserts each error test gives exactly one diagnostic (after E0237 the literal returns the expected type, so no follow-up error).
- Docs: SPEC.md "Literals", D68, `compiler/README.md` Supported.

Proposed SPEC.md text (gate), in Literals:

> - **Arrays:** `[1, 2, 3]`. `[]` needs its type from context, as in `let xs: Array<int> = []`. Arrays have no literal form with a type argument.
> - **Maps:** `Map { "a": 1, "b": 2 }`. A key that is not a literal is written `[key]: value`. The key and value types come from the entries or from context. An empty `Map {}` needs its type from context.
> - **Type arguments on a literal.** `Map<string, int> {}`, `Map<string, int> { "a": 1 }` and `Pair<int, string> { a: 1, b: "x" }` write the type arguments after the name. The literal then has that type without help from context, so `count(Map<string, int> {})` needs no annotation. The compiler checks the arity, that each argument is a type, the map key rules (no float keys, `Equatable` and `Hashable` together) and every key, value and field against the written types. If an expected type also exists, the two must be the same type. A `<` after a capitalized name starts type arguments only when it is followed by a balanced `<...>` of type names, dotted names and tuple types and then `{`; in every other case it is the less-than operator, so `a < b` is unchanged. Function types are not accepted inside the arguments; use an annotation for those. There is no `{K: V}` type and no `#{...}` literal.

## Deviations

- `compiler/internal/narra/selfhost_test.go` is outside the shared list: six cases added to `parseEdgeCases` (data only). `selfhost/src/idcheck.nra` is also outside it: two pattern arities changed.
- Function types are not accepted inside literal type arguments (`Map<string, (int) => int> {}`); use an annotation. Recorded in D68 and README.
- No change requests.

## Stubs left

- None.

## Gotchas

- For selfhost-check-bodies: port `mapLit`/`structLit` checking with `typeArgs`: resolve them as a type, compare with the expected type (E0237), use the result as the literal's type. Any Narra code that matches `MapLit` or `StructLit` needs the extra field.
- `A < B > {` between capitalized names now parses as a literal.
- Narra `Parser.toks` is rewritten by the `>>` split, so the lookahead must run before `litTypeArgs` (it does).
- Adding goldens at the top of a shared dump file shifts every position; use a new file.

## Proposed learnings

- Add AST dump cases in a new `testdata/astdump/*.nra`, not in `exprs.nra` | inserting lines shifts every position in the golden | compiler/testdata/astdump | no

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- `go test ./internal/narra -run 'TestRun|TestErrors|TestASTDump|TestSelfhostParser'` (the brief's verify): pass.

## Size

- Estimated M, actual M.
