# Outcome: patterns

## Built

All code in `compiler/internal/narra/`.

Pattern AST (`ast.go`):

```go
type PatternKind int // PatWildcard, PatName, PatCtor, PatStruct, PatTuple, PatLit

type FieldPattern struct { Pos Pos; Name string; Pattern *Pattern }

type Pattern struct {
	Pos    Pos
	Kind   PatternKind
	Name   string          // constructor name, or the bound name of a PatName
	Enum   string          // qualifier of Shape.Circle
	Args   []*Pattern      // constructor fields or tuple elements; nil for a bare constructor
	Fields []*FieldPattern // PatStruct; `{ y }` is the field y with PatName "y"
	Lit    Expr            // PatLit: IntLit, FloatLit, StringLit, CharLit, BoolLit, or UnaryExpr "-" on a number
	// Set by the checker:
	T       Type     // type of the value this pattern matches
	Variant *Variant // PatName or PatCtor naming a constructor
	Local   *Local   // PatName that binds
	Locals  []*Local // root only: every bound name, in order
}
```

- The old `Wildcard` and `Binds` fields are gone. `Some(v)` is now a `PatCtor` whose `Args` hold one `PatName`.
- `VarStmt.Pattern` is set for `const { a, b: c } = s` and `const [a, b] = t`; `Name` is then empty.
- `StructLit.Spread` holds `s` of `{ ...s, f: v }`. The parser allows one spread, first.
- `ArrayLit.TupleCall` marks `Tuple(a, b)`; `ArrayLit.Tuple` is set by the checker when the literal builds a tuple (also `[a, b]` with a tuple expected).
- `StructType.Tuple`: `Tuple<A, B>` is a prelude `StructType{Name: "Tuple", Module: "", Args: elems, Fields: "0", "1", ...}`, interned by `c.tupleOf(elems)` (`check_patterns.go`, map `Checker.tuples`). C name `S_Tuple__int__string`, fields `f_0`, `f_1`. The struct paths (layout, `orderTypes`, retain and release, eq, hash, cmp) handle it unchanged. `substitute` has a tuple case. `tupleOf` registers trait facts for `typeKey{"", "Tuple"}`: derives `Ordered` and `Show`, automatic `Equatable` and `Hashable`.

Checker (`check_match.go`, `check_patterns.go`):
- `c.resolvePattern(p, t)` checks a pattern against a type and fills `T`, `Variant`, `Local`, and `Locals` on the root. A `PatName` is a constructor when `t` is an enum with that variant; otherwise it binds, and an uppercase name is an error. Constructors in patterns resolve through the subject's type, never through the constructor ranks, so a user `Less` cannot capture `Ordering.Less` and `Message.Text` works on an imported enum (`namesEnum`).
- `c.match` accepts any subject type. Exhaustiveness uses the usefulness algorithm (`missingCases`), with witnesses: enums, bool, structs and tuples have finite constructors; other types need a wildcard or binding. Arms with guards do not count. Messages: `match does not cover Some(Circle(_, false)), Some(Square); add the cases or a `_` arm`, or `match does not cover every int value; add a `_` arm`.
- `irrefutable(p)` and `argsIrrefutable(p)`. Destructuring needs an irrefutable pattern. `if (x is Some(1)) ... else` no longer narrows x to `None` in the else branch; it does only when the constructor's arguments are irrefutable (`check_stmt.go`).
- `c.isExpr`, `c.destructure`, `c.spreadLit`, `c.tupleType`, `c.tupleLit`.

Emitter (`emit_match.go`, `emit_patterns.go`):
- The subject is copied to a C temporary (`tN`) that lives for the whole match, `if`, or `const` statement. `e.patternCond(p, path)` builds a `&&` chain of tests on C paths into it: `t3.tag == 0 && t3.u.v_Some.f_value.tag == 1 && lstr_eq(t3.u.v_Some.f_value.u.v_Node.f_label, (&L2))`. Tag tests come before reads of their union member, so `&&` short-circuit keeps reads valid.
- `e.bindPattern(p, path)` declares each bound local as `T v = path;` plus a retain, and adds it to the current scope.
- `matchArms` emits per arm: `if (!done) { bool c = <cond>; if (c) { <bindings> if (<guard>) { done = true; <body> } <release bindings> } }`.
- `ifIs` holds the subject as a scope local, computes the condition, then binds in the then-branch.
- `IsExpr` outside an `if` condition: subject temporary plus `patternCond`.

How reference counts flow:
- The match subject is borrowed (or an owned temporary of the enclosing statement). An `if ... is` subject is consumed into a held local released at the end of the `if`.
- Every bound name owns one reference (retain at bind, release when the arm's or branch's scope pops, or on `return`, `break` and `continue` through the normal unwind). A guard that fails releases the bindings.
- Destructuring: the value goes into a temporary; each name retains its part; an owned value is released at the end of the statement.
- Spread: given fields are consumed (owned), untouched fields are copied from the source and retained, so the new struct owns all its fields. An owned source is released at the end of the statement.
- Tuple literals consume each element, like struct literals.

Tests: samples `compiler/testdata/patterns_nested.nra`, `patterns_structs.nra`, `patterns_tuples.nra`, `patterns_generic.nra` (with `.out`). Package `compiler/testdata/packages/patterns` (matches on `pp/model`'s enum, destructures and spreads its struct, takes a tuple from it, matches `Ordering` while an imported enum has `Less`). Error tests `compiler/testdata/errors/patterns_*.nra`: shorthand, spread type, spread order, nested exhaustiveness, int exhaustiveness, refutable destructuring, name bound twice, tuple arity.

## Deviations

- Outside the brief's paths: `check.go` (one field, `Checker.tuples`) and `emit_traits.go` (tuple case in `showDef`, writes `Tuple(1, "a")`). New files `check_patterns.go` and `emit_patterns.go` instead of growing the shared files, as the orchestrator asked. `compiler/README.md` updated.
- Language choices recorded in `compiler/README.md` "Differences from the spec": the constructor-or-binding rule for names, `Tuple(...)` parsed as syntax, `match` and `is` on any value, `let` destructuring, destructuring needs irrefutable patterns.
- `IsExpr` emission now uses a subject temporary, so the C for existing `is` expressions changed; program behavior is the same.

## Stubs left

- No `for (const [i, x] of ...)` destructuring (needs `Array.entries`, generators).
- No tuple element access (`t.0`). No typed struct patterns (`Point { x }` as a pattern). No unreachable-arm warning.
- Opaque structs (opaque-distinct) must reject struct patterns, destructuring and spread of another module's opaque struct in `resolvePat` (PatStruct) and `spreadLit`.

## Gotchas

- `Tuple` is a parser keyword-like name only when followed by `(`; as a type it resolves in `resolveType` before user types.
- `[a, b]` with no expected type is still an array. Generic parameters get no expected type (`genericCall` passes nil for open types), so pass `Tuple(a, b)` there.
- Match arms are no longer a single `if (!done && tag)`: each arm opens two C blocks. Closures that capture pattern locals see ordinary `Local`s with `CName` `v_<name>_<n>`.
- `Pattern.Locals` is only set on the root pattern. Nested `Local`s are on `Pattern.Local`.

## Proposed learnings

- Emit pattern tests as `&&` chains with each tag test before the reads of its union member | the reads are only valid after the tag check | compiler/internal/narra | yes, ASan catches union misreads in samples
- Bound pattern names retain at bind and release at scope pop, never borrow from the subject | the subject temporary can be released before the arm ends | compiler/internal/narra | yes, the leak check

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (all samples, including the new ones, under ASan, UBSan and the leak check, plus release builds; the new package test in debug, release and test mode; all error tests).
- `narra build` plus `NARRA_CHECK_LEAKS=1` on `patterns_generic.nra` by hand: exit 0.

## Size

- Estimated L, actual L.
