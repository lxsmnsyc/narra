# Outcome: compound-assign

## Built

- `Stmt.CompoundAssignStmt(pos, target, operator, value)` in `selfhost/src/ast.nra`; cases in the checker (`stmts.nra`: `checkAssignment` takes an operator, `checkCompoundOperand`), effects (`alloc` for string joins and Map targets), the emitter (`emitstmts.nra`: `emitCompoundAssignment`, `emitMapCompound`; `emitexprs.nra`: `emitArithmetic` split out of `emitBinary`), `fmt`, `astdump`, `typesdump`, ids, query sites and trivia.
- Parser (`parser.nra`): `parseCompoundTail`, `isCompoundOperatorAhead`, `isIncrementAhead`, `Parser.statementStart`. The c-style-for step accepts `i += 1` and `i++`.
- Lexer: `x++` / `x--` at the end of a line ends the statement.
- Codes: E0962 (operand type or operator does not fit the target), E0963 (`++`/`--` on a non-integer), E0964 (prefix `++x`, or a compound assignment used as a value). E0034 removed.
- Tests: `testdata/compound_assign.nra/.out`, `testdata/errors/compound_assign_*` (six), `testdata/astdump/compound_assign`, new cases in `testdata/c_style_for`, golden `testdata/parse/compound.ast`.
- SPEC.md (section 2 Bindings, Operators, Differences table), D76 amendment, guide chapters 2 and 11 (their "error" examples).

## Deviations

- The brief says reading `counts[key]` panics on a missing key; it returns `Option<V>`. SPEC states that only `counts[key] += e` panics.
- The value is evaluated before the target place (as `=` does), not after.
- `+%=`, `-%=` and `*%=` also work.
- `--x` used to parse as a double negation; it is now E0964.
- No panic sample for the Map miss: the panics suite list is in `selfhost/tests/src/panics.nra`, outside this node's paths. Add a `PanicSample` for a `counts["b"] += 1` program ("map key not found").
- The inbox entry of 2026-11-07 (SPEC and `+=`) is answered; the orchestrator removes it.

## Gotchas

- Do not use the syntax in `selfhost/src` before the seed is regenerated.
- A parser path that parses an expression statement must set `parser.statementStart = parser.index` first, or `+=` gives E0964.
- `x++` is recognised only before `)`, `}` or the end of a line; `a++ + b` is not supported.

## Verification

- Targeted suites with a compiler built from the branch; `sh scripts/test.sh` (see the final report).

## Size

- Estimated M, actual M.
