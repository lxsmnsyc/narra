# Outcome: ts-habit-fixes

## Built

Go: `compiler/internal/narra/parser.go` (`headerLine`, hooks into `decl`, `primary` and the `for` statement), `diagnostic_fixes.go` (`habitDecl`, `habitExpr`, `speculateExpr`, `failForIn`, `quoteFix`, `lengthFix`, `stdAliasPaths`), `lexer.go` (quote fix), `check_expr.go`, `check_types.go`, `check_call.go` (fixes), `diagnostic_codes.go` (E0923-E0928 and their `explain` texts).
Selfhost: `selfhost/src/parser.nra` (`reportHabitWord`, `isThrowOperand`, class and interface cases). Same messages and positions. Fixes exist only in Go, like all earlier fixes.
Tests: `testdata/diagnostics/ts_*.nra` (nine cases, fix applied must compile, run by `TestHabitFixes`), `testdata/errors/ts_{class,interface,new,try,throw,this}.nra`, `TestHabitExplain`, and `TestHabitCases` (inline programs, including every shape that must get no fix; each offered fix is applied and re-checked clean).

| TypeScript habit | Code | Message | Fix |
| --- | --- | --- | --- |
| `class A {}` | E0923 (new) | there is no `class`; use a `struct` and functions in a module | `class` to `struct`, only when the body is `name: type` lines |
| `interface A {}` | E0924 (new) | there is no `interface`; use a `struct` for data or a `trait` for behavior | `interface` to `struct`, same condition (a method gives no fix) |
| `new A()` | E0925 (new) | there is no `new`; write a struct literal such as `Name { field: value }`, or call a function | none (the parser cannot tell a struct from a function) |
| `try { } catch` | E0926 (new) | there is no `try` or `catch`; return a `Result` and check it with `is Err` or `match` | none (no single exact edit) |
| `throw x` | E0927 (new) | there is no `throw`; return `Err(value)`, or call `panic` for a bug | `return Err(x)`, only for `throw` that starts a statement in a function declared to return `Result` |
| `this` | E0928 (new) | there is no `this`; pass the value as a parameter | none; also reported at a declaration such as `const this` or a parameter named `this` |
| `for (const k in m)` | E0031 | use `of`; `for ... in` is not part of the language | none (the parser cannot tell a map from an array) |
| `number` | E0113 | there is no `number`; use `int` or `f64` | two fixes: `int`, `f64` |
| `boolean` | E0114 | use `bool` | `bool` |
| `'abc'` | E0002 | a character literal holds exactly one character; use double quotes for strings | swap the quotes (no escapes or `"` inside) |
| `s.length` | E0208 | there is no .length; use Array.length(x) or Str.length(x) | `Str.length(s)`, `Arrays.length(xs)` or `Maps.size(m)`, adds the import |
| `Arrays.push(...)` without import (and `Maps`, `Str`, `Int`, ...) | E0308 | unknown module "X"; import it first ... | add `import * as X from "std/..."` |

## Deviations

- Rows E0926 (try/catch) and E0928 (this) have no fix: no single exact edit exists. The brief asks for "a fix where one edit is right".
- Rule from review: a fix is right or absent. `new`, `for in`, a class or interface with methods, and `throw` outside a Result function have no fix.
- E0114 (`boolean`) also gets a fix; it was not in the brief but is the same kind of habit.
- `diagnostic_test.go` gained `TestHabitExplain`. No other path outside the brief was edited.

## Stubs left

- None.

## Gotchas

- The six words are ordinary identifiers to the lexer; only the parser positions listed above report them.
- Parser fixes needing imports use `Parser.imports` and `headerLine`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (334 s, includes TestSelfhost* and TestSelfhostBootstrap).

## Size

- Estimated M, actual M.
