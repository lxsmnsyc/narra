# Log: let-narrowing

- Nothing existed: `narrowable` refused every `Mutable` local.
- Go: `narrowable` accepts any local. `endNarrowing(l)` deletes the narrowing of a `let` or `mutable` local from every scope. Called after an assignment, after a call that passes the local as `mutable`, and by `endLoopNarrowing` before a `while` or `for` is checked. The loop pre-scan (`assignedInBlock`, `assignedInExpr`) collects root names assigned or passed as `mutable` in the condition and body, matched by name.
- selfhost: same in `patterns.nra` (`endNarrowing` writes -1 under the narrow key in every scope), `stmts.nra` (assignment, loops, scanners) and `exprs.nra` (`checkCallEndingNarrowing`).
- Arrow function bodies are not scanned: a closure cannot capture a `let`.
- Error text unchanged; no new diagnostic code.
