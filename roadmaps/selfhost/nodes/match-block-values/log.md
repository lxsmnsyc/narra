# Log: match-block-values

- Nothing existed: `check_match.go` reported "a match used as a value needs a single expression in each arm".
- Design: no AST change. In value position the checker and emitter treat `MatchArm.Block` as a block expression through `blockBody` (Go `check_stmt.go`, `emit.go`; selfhost `checkBlockBody` in `stmts.nra`, `emitBlockBody` in `emitmatch.nra`). A new `Walker.checkBlockBody` trait method reaches it from `match.nra`.
- `return match (...)` with a tail call (`returnMatch`) also handles block arms now (it would have read a nil `Body`).
- Walkers: Go `usesLocal` (`check_call.go`) and `mutableRoots` (`check_closure.go`) now visit arm blocks; selfhost `armsUseLocal` and `collectArmRoots` likewise. The exclusivity walker is `collectExpressionAndArms`/`visitExpr`, which already visited arm blocks.
- Narra gotcha: `f(mutable emitter, typeOf(emitter, x))` is rejected; compute the type first.
- Existing golden `typesdump/patterns/err_match` held the old error; its `blockInValue` arm now returns a string so the case still fails, with a new message.
