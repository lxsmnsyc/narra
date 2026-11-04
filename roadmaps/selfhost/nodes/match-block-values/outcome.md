# Outcome: match-block-values

## Built

- A `match` arm written `pattern => { ... }` in value position is a block expression (D75): `return v` gives the arm's value and ends the arm, `break`/`continue` cannot leave it (E0932), every path needs `return` (E0931). A `match` in statement position is unchanged.
- Go: `blockBody` in `check_stmt.go` and `emit.go` (the old `blockExpr` calls it); `check_match.go` and `emit_match.go` use it for arm blocks; `returnMatch` in `emit_call.go` handles block arms; `usesLocal` and `mutableRoots` visit arm blocks.
- selfhost: `checkBlockBody` (`stmts.nra`, trait method `Walker.checkBlockBody` in `check.nra`, `body.nra`), `emitBlockBody` and `emitReturnBlockArm` (`emitmatch.nra`), `armsUseLocal` (`exprs.nra`), `collectArmRoots` (`closures.nra`).
- Tests: `compiler/testdata/match_block_values.nra` + `.out`; `compiler/testdata/errors/match_block_values_{no_return_path,arm_types,break_across}.nra`.
- SPEC.md "Pattern matching" describes the rule.

## Deviations

- No AST change; the arm keeps `Block`, and value position is decided by the checker (`IsValue`).
- Edited outside the owned and shared lists: `compiler/testdata/typesdump/patterns/err_match.{nra,dump}` (its old error no longer exists), `selfhost/src/check.nra`, `body.nra`, `stmts.nra`, `exprs.nra`, `closures.nra`, `emitexprs.nra` (all under `selfhost/src/**`, shared), and `compiler/internal/narra/check_call.go`, `check_closure.go`, `check_stmt.go`, `emit.go`, `emit_call.go` (not in the shared list).

## Stubs left

- None.

## Gotchas

- Error for an arm of another type is reported at the offending `return` value when the result type is already known, and at the block's `{` when an expression arm comes after a block arm.
- Any new walker over match arms must visit `arm.Block`.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: see PR.

## Size

- Estimated S, actual S.
