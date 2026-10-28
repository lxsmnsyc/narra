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
- `VarStmt.Pattern` is set for `const { a, b: c } = s` and `const #[a, b] = t` (also with `let`); `Name` is then empty.
- `StructLit.Spread` holds `s` of `{ ...s, f: v }`. The parser allows one spread, first.
- The lexer has a `#[` token. `ArrayLit.IsTuple` marks `#[a, b]`; `ArrayLit.Tuple` is set by the checker. A tuple literal needs no expected type; an expected tuple of the same length guides its elements. `[a, b]` is always an array. `TypeExpr.Tuple` marks the type `#[A, B]`, whose `Args` are the element types.
- `StructType.Tuple`: `#[A, B]` is a `StructType{Name: "Tuple", Module: "", Tuple: true, Args: elems, Fields: "0", "1", ...}` that no source name refers to, interned by `c.tupleOf(elems)` (`check_patterns.go`, map `Checker.tuples`). C name `S__Tuple__int__string` (no module path maps to a segment starting with `_`, so this cannot equal a user struct; `testdata/Tuple.nra` would collide with `S_Tuple__...`), fields `f_0`, `f_1`. `String()` prints `#[int, string]`. The struct paths (layout, `orderTypes`, retain and release, eq, hash, cmp) handle it unchanged. `substitute` has a tuple case. `tupleOf` registers trait facts for `typeKey{"", "Tuple"}`: derives `Ordered` and `Show`, automatic `Equatable` and `Hashable`.

Checker (`check_match.go`, `check_patterns.go`):
- `c.resolvePattern(p, t)` checks a pattern against a type and fills `T`, `Variant`, `Local`, and `Locals` on the root. A `PatName` is a constructor when `t` is an enum with that variant; otherwise it binds, and an uppercase name is an error. Constructors in patterns resolve through the subject's type, never through the constructor ranks, so a user `Less` cannot capture `Ordering.Less` and `Message.Text` works on an imported enum (`namesEnum`).
- `c.match` accepts any subject type. Exhaustiveness uses the usefulness algorithm (`missingCases`), with witnesses: enums, bool, structs and tuples have finite constructors; other types need a wildcard or binding. Arms with guards do not count. Messages: `match does not cover Some(Circle(_, false)), Some(Square(_)); add the cases or a `_` arm`, or `match does not cover every int value; add a `_` arm`. A variant with fields prints as `Name(_)` even when every field is a wildcard. After a pattern error the arm's guard and body are not checked and the exhaustiveness check is skipped; names inside a failed pattern are bound with type void (`bindRest`), so no `unknown name` follows.
- `irrefutable(p)` and `argsIrrefutable(p)`. Destructuring needs an irrefutable pattern; `const Only(s) = u` parses (a name followed by `(` or `.` after `const`/`let`) and works when the enum has one variant. `if (x is Some(1)) ... else` no longer narrows x to `None` in the else branch; it does only when the constructor's arguments are irrefutable (`check_stmt.go`).
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

Tests: samples `compiler/testdata/patterns_nested.nra`, `patterns_structs.nra`, `patterns_tuples.nra`, `patterns_generic.nra`, and `Tuple.nra` (module "Tuple" using tuples, the C name collision case), all with `.out`. `patterns_tuples.nra` passes `#[..]` straight to generic functions and declares its own `struct Tuple`. Package `compiler/testdata/packages/patterns` (matches on `pp/model`'s enum, destructures and spreads its struct, takes a tuple from it, matches `Ordering` while an imported enum has `Less`). Error tests `compiler/testdata/errors/patterns_*.nra`: shorthand, spread type, spread order, nested exhaustiveness, int exhaustiveness, refutable destructuring (struct and constructor), name bound twice, tuple arity, one-element tuple literal (`#[x]`), one-element tuple type (`#[int]`), `Tuple<int, int>`, array literal where a tuple is expected.

## Deviations

- Outside the brief's paths: `check.go` (one field, `Checker.tuples`) and `emit_traits.go` (tuple case in `showDef`, writes `#[1, "a"]`). New files `check_patterns.go` and `emit_patterns.go` instead of growing the shared files, as the orchestrator asked. `compiler/README.md` updated.
- Language choices recorded in `compiler/README.md` "Differences from the spec": the constructor-or-binding rule for names, `match` and `is` on any value, `let` destructuring of structs, destructuring needs irrefutable patterns.
- Tuple syntax `#[...]` for literals, patterns and types (user decisions during review). `SPEC.md` edited in section 2 (destructuring line, `for (const #[i, item] of ...)`, the Tuples literal line), section 3 (prelude list no longer has `Tuple`), section 4 (Tuples under Composite types, the "type syntax uses names" line) and section 8 (`const #[sender, receiver] = Channel.create`, its `#[Sender<T>, Receiver<T>]` result). `DECISIONS.md`: D23 notes tuples as the exception, D44 example now `const #[a, b] = f()`, and new decision D53 "Tuple literals and types use `#[...]`" (D51 and D52 belong to the closures PR). `lexer.go` gained the `#[` token and `check_traits.go` `typeExprString` prints tuple types (both outside the brief's paths).
- The C for `match` and `is` changed shape: the `match` subject and the `is` subject are held in a temporary, and each arm computes its condition into a `bool` before testing it. Program behavior is the same.

## Stubs left

- No `for (const [i, x] of ...)` destructuring (needs `Array.entries`, generators).
- No tuple element access (`t.0`). No typed struct patterns (`Point { x }` as a pattern). No unreachable-arm warning.
- Opaque structs (opaque-distinct) must reject struct patterns, destructuring and spread of another module's opaque struct in `resolvePat` (PatStruct) and `spreadLit`.

## Gotchas

- `Tuple` is a plain name again: a program can declare `struct Tuple`. `Tuple<A, B>` without such a declaration gives "there is no Tuple type; write a tuple type as #[A, B]".
- `#[` counts as an opening bracket in the lexer, so line breaks inside a tuple literal do not end the statement. Tuple patterns are written `#[p, q]`; `[` starts no pattern.
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
