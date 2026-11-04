# Outcome: let-narrowing

## Built

- A `let` (and a `mutable` parameter) narrows after an `is` test like a `const`: `if (x is None) { return }` then `x.value`, `is Err(e)` early return then `r.value`, and inside `if (x is Some(v))`.
- Statements that end the narrowing of `x`, in every enclosing scope and not only the current one:
  1. an assignment whose target is rooted at `x` (`x = ...`, `x.f = ...`, `x[i] = ...`), after its value is checked;
  2. a call that passes `x` (or a field or element of it) as `mutable`, after the call is checked;
  3. the start of a `while` or `for` whose condition or body contains 1 or 2 for `x`: the narrowing ends before the loop is checked and stays ended after it.
- Not ending it: reading `x`, passing it by value, calls without `mutable x`, assignments to other names. A new `is` test narrows again.
- Go: `compiler/internal/narra/check_stmt.go` (`narrowable`, `endNarrowing`, `endLoopNarrowing`, `assignedInBlock`, `assignedInStmt`, `assignedInExpr`, `rootName`), `check_expr.go` (call case).
- selfhost: `selfhost/src/patterns.nra`, `stmts.nra`, `exprs.nra`.
- Tests: `compiler/testdata/let_narrowing.nra` (+ `.out`), `compiler/testdata/errors/let_narrowing_{assign,mutable_arg,loop}.nra`.

## Deviations

- `compiler/testdata/typesdump/patterns/err_narrowing.dump` (outside the owned paths): removed the two lines for `mutableNotNarrowed`, which tested that a `let` is not narrowed. That function now compiles.

## Stubs left

- None.

## Gotchas

- Ending is conservative: an assignment in an `if` branch ends the narrowing after the `if` too. The loop scan matches names without scopes and does not look into arrow bodies.
- `endNarrowing` only affects `let` and `mutable` locals; `const` narrowing is unchanged.
- In Narra, `test` is a keyword, and a call that takes `mutable checker` cannot have another call using `checker` as an argument.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass.
- `go test -count=1 ./internal/narra -run TestSelfhost`: pass (includes the bootstrap fixed point).

## Size

- Estimated M, actual M.
