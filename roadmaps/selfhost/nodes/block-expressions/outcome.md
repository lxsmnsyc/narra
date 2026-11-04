# Outcome: block-expressions

## Built

- AST: `BlockExpr` (`ast.go`, `clone.go`, `astdump.go` kind `blockexpr`, `typesdump.go`); selfhost `Expr.BlockExpr(id, pos, block)` in `ast.nra`, `astdump.nra`, `typesdump.nra`, `idcheck.nra`, `check.nra`.
- Parser (`parser.go` `blockExpr`, `ifExpr`, `structLitAhead`; `parser.nra` `parseBlockExpr`, `parseIfExpr`, `isStructLitAhead`): `primary` takes a `{` or `if` as an expression. An `if` as a value is desugared in the parser into a `BlockExpr` that holds one `if` statement whose two branches each hold `return <BlockExpr>`. The checker and emitter have no `IfExpr`. The `else` is required (`expected "else", found line break`), `else if` chains nest.
- Checker (`check_stmt.go` `blockExpr`, `blockReturn`, `endsInReturn`, `loopExit`; `stmts.nra` `checkBlockExpression`, `checkBlockReturn`, `endsInReturn`, `checkLoopExit`): a stack `Checker.blocks` (`blockCtx` / `BlockContext`) is the return target. It is cleared in function bodies, arrows (`check_closure.go`, `closures.nra`) and constant checks, and saved in `startBody`.
- Emitter (`emit.go` `blockExpr`, `blockReturn`; `emitexprs.nra` `emitBlockExpression`, `emitstmts.nra` `emitBlockReturn`, `emitstate.nra` `unwindBlock`): `T rN; { ... } narra_block_N:;`. A `return v` stores the owned value in `rN`, releases the temporaries and scopes opened since the block started, and writes `goto narra_block_N;`. A result with heap data is an owned temporary of the enclosing statement, like a `match` value. `Emitter.blocks` is cleared per function and saved around closure bodies.
- Codes: E0931 path without `return value` (also a bare `return`, or a `return` of a void call, inside a block), E0932 `break`/`continue` leaving a block expression, E0933 block with no `return` (`const o = { a }`).
- Tests: `compiler/testdata/block_expressions.nra` + `.out`; `compiler/testdata/errors/block_expressions_{no_return_path,break_across,object_literal,bare_return,branch_types,if_no_else}.nra`; `compiler/testdata/astdump/block_expressions.nra` + `.ast`.
- SPEC.md section 2 ("Block expressions and `if` as a value") and a "Differences from TypeScript" row; `compiler/README.md` "Supported".

Value positions (everything `primary` reaches). A `{` or `if` there is a block expression:

- the initializer of `const`/`let`, a top-level `const` value (then rejected by the constant rules), the right side of `=`
- a call argument and a callee, an operand of a unary or binary operator, an `is` subject, an index and a field target
- a `return` value, a `${...}` in a template
- an element of an array or tuple literal, a struct field value, a spread source, a map key or value
- a condition (`if`, `while`), a `for ... of` iterable, a `match` subject, a guard, an arrow's expression body
- a branch of an `if` as a value, and the single-expression body of a `match` arm (`=> expr`)

Not value positions (still statements or bodies): the start of a statement (`{` is still error "a block must belong to a statement"; `if` is `IfStmt`), the body of `if`/`else`/`while`/`for`, a block-bodied arrow, a function body, and a `match` arm written `=> { ... }` (that is `MatchArm.Block`, unchanged).

For match-block-values and ternary:

- Reuse `BlockExpr` and its `Checker.blocks` target. To give a `match` arm block a value, make the arm's block a `BlockExpr` (or push a `blockCtx` around it) so `return` ends the arm; the emitter's `blockExpr` already makes `rN`, the label and the unwinding, and `matchValue` can wrap the same way.
- `parseIfExpr` shows the desugar: build `Block{ ReturnStmt(value) }` around each branch. A ternary `c ? a : b` can parse to `BlockExpr{ IfStmt(c, {return a}, {return b}) }` with no checker or emitter change; only the mismatch error position (the second branch's expression) and the missing-`:` parse error are new.
- A `return` inside a block expression is never a tail call (`findTailCalls` does not scan expressions), so `return n < 0 ? f(n + 1) : n` stays a plain call unless ternary teaches `findTailCalls` about it.

## Deviations

- Single-expression `if` branches are not supported: a branch is a block (`{ return 1 }`). The brief's "single-expression branches" is read as a one-statement block.
- A `{` in value position stays a struct literal when it starts `{}`, `{ ...`, or `{ name:` not followed by `for`/`while` (context-typed struct literals existed already). Only other `{` become block expressions. This keeps every current program valid.
- Edited outside the owned and shared lists: `compiler/internal/narra/astdump_test.go` (`astKinds` gets `blockexpr`), `compiler/testdata/astdump/block_expressions.{nra,ast}` (the dump test needs a golden with every kind), and selfhost files `body.nra`, `emitclosures.nra`, `emitconsts.nra`, `emitstate.nra`, `emitexprs.nra`, `emitstmts.nra`, `effects.nra`, `idcheck.nra`, `typesdump.nra`, `exprs.nra`, `closures.nra` (Walker trait and per-function state live there). Go: `typesdump.go`.
- No missing-return check exists for functions, so `endsInReturn` is new. It accepts a final `return`, `panic`, a block, an `if`/`else` where both end in `return`, and a `match` statement whose arms all end in `return` or `panic`. A loop never counts (`while (true)` needs a `return` after it).

## Stubs left

- None.

## Gotchas

- `mutableRoots` and `usesLocal` (Go and selfhost) do not look into block expressions, as they do not look into match arm blocks. A `mutable` argument inside a block expression is not seen by the "passed as mutable and also the callee" rule.
- E0931 and E0933 are reported at the block's `{` (the `if` keyword for an `if` value); the error for different branch types is at the second branch's `{`.
- An error inside a block (for example `const o = { a }` also reports "this expression does nothing" for `a`) appears with the block errors.
- The result variable is declared without an initializer (`T rN;`), like `match` values; every path assigns it before the `goto`.
- Narra source: a call that takes `mutable checker` cannot use `checker` in another argument, so compute the value first (see `checkBlockExpression`).

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (about 171 s), including `TestSelfhostBodies`, `TestSelfhostEmit`, `TestSelfhostBootstrap`, `TestASTDump`, `TestEveryFormatHasCode` and the new samples and error tests.
- `narra run compiler/testdata/block_expressions.nra` with `NARRA_CHECK_LEAKS=1`: the output equals the `.out`; the sample also runs in `TestPrograms` under ASan, UBSan and the leak check.

## Size

- Estimated M, actual M.
