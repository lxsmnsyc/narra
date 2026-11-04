# Outcome: ternary

## Built

- Lexer: `?` is a punctuation token (`lexer.go` `puncts`, `lexer.nra` `PUNCTS`).
- Parser: `expr` (Go) and `parseExpr` (selfhost) parse `binary` and then an optional `? a : b`, right-associative and looser than `||`. The result is a `BlockExpr` holding one `IfStmt` whose branches each `return` one value, the same desugar as `if` in value position. No checker, emitter, walker, clone or dump change was needed. The parser remembers which block expressions came from a ternary (`Parser.ternaries`, selfhost `Parser.ternaryIds`).
- Tail calls: `return c ? a : b` is parsed into an `if` statement (`returnedTernary`, `convertReturnedTernary`), nested ternaries too, so the existing tail-call pre-pass sees `return f(...)` in a branch. `climb(-5000000)` and `countdown(5000000, 0)` in the sample run as loops.
- Code E0934: a missing `:` ("a conditional expression needs `:` and a second value, as in `cond ? a : b`"), in `diagnostic_codes.go`.
- Tests: `compiler/testdata/ternary.nra` and `.out`; `compiler/testdata/errors/ternary_{condition_not_bool,branch_types,missing_colon}.nra`.
- E0929 (mutable-read-conflict) needed no code: a ternary is a block expression holding an `if`, and the use scan already treats the two branches of an `if` as exclusive. Tests: `errors/ternary_mutable_conflict.nra` (`x + (c ? bump(mutable x) : 0)`) and `ternary_mutable.nra` (`c ? bump(mutable x) : x` is legal).
- Docs: SPEC.md section 2 (block expressions list), DECISIONS.md D75 line, `compiler/README.md` "Supported".

## Deviations

- Edited `DECISIONS.md` and `compiler/README.md`, which are not in the brief's lists (the brief asks for the D75 line; README is the project rule).
- No `Ternary` AST node; this is why `clone.go`, `astdump.go`, `check_*.go`, `emit_expr.go` and the selfhost walkers are untouched.
- A non-`bool` condition reports the existing `expected bool, found int` (the same message `if` gives); no new text.
- In `return c ? a : b`, a branch type error is the function's return type error (it is an `if` of returns), not a branch mismatch.

## Stubs left

- None.

## Gotchas

- Only a ternary that is the whole value of a `return` is rewritten to an `if`. `return 1 + (c ? f(n) : 0)` is not a tail call, and a ternary elsewhere never is.
- The second operand parses with `expr`, so `a ? b ? c : d : e` nests as in TypeScript.
- `{ name: value }` after `?` is still read as a struct literal by `structLitAhead`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (about 300 s), including TestSelfhostBootstrap and the new sample and error tests.

## Size

- Estimated S, actual S.
