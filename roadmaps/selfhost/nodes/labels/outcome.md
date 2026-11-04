# Outcome: labels

## Built

- Go: `label` on `WhileStmt`, `ForStmt`, `BreakStmt`, `ContinueStmt` (`ast.go`, `clone.go`, `astdump.go` writes `label="x"` only when set). Parser: `Parser.labeledStmt`, `loopExitLabel` (`parser.go`). A labeled loop takes the label's position as its `Pos`.
- Checker (`check_stmt.go`): `Checker.labels`, `enterLoop`, `leaveLoop`, `loopExit`; reset in closures and constant evaluation.
- Emitter: `loopInfo`, `enterLoop`, `leaveLoop`, `labeledExit`, `unwindToLoop`, `continueLabel`, `breakLabel` (`emit.go`); used by `while`, range, array and map loops. `break outer` releases everything down to the target loop and writes `goto narra_break_L_C;` after the loop; `continue outer` writes `goto narra_continue_L_C;` to a label at the end of the body. A C label is written only when used.
- Codes: E0916 no enclosing loop is labeled, E0917 label already used by an enclosing loop, E0918 label must be on a for or while loop.
- selfhost: same changes in `ast.nra`, `astdump.nra`, `parser.nra`, `stmts.nra`, `check.nra`, `closures.nra`, `emitstate.nra`, `emitstmts.nra`, `emitmaps.nra`, `emitclosures.nra`, `emitconsts.nra`; pattern arity updates in `effects.nra`, `idcheck.nra`, `typesdump.nra`, `emittail.nra`.
- Tests: `compiler/testdata/labels.nra` and `.out`; `compiler/testdata/errors/labels_unknown.nra`, `labels_reused.nra`, `labels_not_loop.nra`.

## Deviations

- Edited selfhost files outside the brief's shared list (`check.nra`, `closures.nra`, `effects.nra`, `idcheck.nra`, `typesdump.nra`, `emittail.nra`, `emitstate.nra`, `emitmaps.nra`, `emitclosures.nra`, `emitconsts.nra`) because `Stmt` pattern arity and checker and emitter state live there. Go: `check_closure.go`, `check_const.go`, `check.go`, `emit_closure.go`, `std_map.go` for the same reason.

## Stubs left

- None.

## Gotchas

- `Stmt.WhileStmt(pos, label, condition, body)`, `ForStmt(id, pos, label, name, iterable, body)`, `BreakStmt(pos, label)` and `ContinueStmt(pos, label)` gained a field. Other branches that add patterns on them must update arity.
- Narra has no `Arrays.pop`; selfhost rebuilds the label and loop stacks.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (full suite 229 s).

## Size

- Estimated M, actual M.
