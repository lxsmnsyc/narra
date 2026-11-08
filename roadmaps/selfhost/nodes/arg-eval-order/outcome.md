# Outcome: arg-eval-order

## Built

- Both emitters evaluate call arguments in written order. When a later argument may assign a variable (a block expression, a `match`, a block-bodied arrow, a call of a closure value, a local closure passed on), each earlier non-`mutable` argument that is not a literal is evaluated first into a temporary. A name read of a reference-counted value is retained in that temporary and released at the end of the statement.
- Go: `holdEarlierArgs` and `mayWrite` in `compiler/internal/narra/emit_call.go`, called at the start of `call`; `Emitter.held` in `emit.go`; `expr` in `emit_expr.go` returns a held value once. One hook covers user calls, closure calls, trait calls, constructors and every builtin path.
- Selfhost: `holdEarlierArguments` and `mayWrite` in `selfhost/src/emitexprs.nra`, called from `emitCall`; `heldArguments` and `heldPending` in `emitstate.nra`.
- Sample `compiler/testdata/arg_eval_order.nra` (+ `.out`): block argument, a call with a side effect after a read of the same `let`, calls with printed side effects, a self tail call, and a std builtin call (`Str.slice`).
- `SPEC.md` section 2 (Functions) states the rule. Inbox item (2) of the PR #96 entry is removed.
- Self tail calls already hoisted arguments in index order (`consume` into a temporary); no change was needed, and the sample covers them.
- `mutable` arguments: a later argument that uses the borrowed value is already rejected ("passed as mutable in this call, so no other argument can use it"); no temporary.

## Proof

- base: 6e42877, `cd compiler && go run ./cmd/narra run testdata/arg_eval_order.nra` prints `1, 55, first, second, 12, 101001, wx`; expected `3, 15, first, second, 12, 101001, ab`.
- branch: same command prints `3, 15, first, second, 12, 101001, ab`. `TestPrograms` passes (ASan, leak check), and the self-hosted emit of the sample equals Go's.

## Deviations

- `roadmaps/selfhost/inbox.md` edited (not an owned path) to remove item (2), as the brief asks.

## Stubs left

- None.

## Gotchas

- Generated C changed for samples with a block-bodied arrow or block argument after an earlier name read: `block_arrow_args`, `block_expressions`, `closure_borrow`, `closures`, `for_of_sources_nested` (an extra temporary, and a retain/release pair for reference-counted reads). All 76 other top-level samples are byte-identical (checked by emitting before and after).
- `mayWrite` is conservative: any node it does not know counts as a writer. A later node that adds an expression type should add a case if the type cannot assign.
- Named arguments (D87) can call `holdEarlierArgs` with the arguments in written order, then reorder.
- Held values are keyed by expression node (Go) or node id (selfhost); a selfhost entry is marked taken with empty text because maps have no delete.

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestPrograms`: pass.
- `cd compiler && go vet ./... && go test -timeout 45m ./...`: pass (537 s).

## Size

- Estimated M, actual M.
