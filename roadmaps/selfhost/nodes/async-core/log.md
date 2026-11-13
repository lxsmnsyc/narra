# Log: async-core

## Proof

- base feat/batch-12 d9d1bf12: `narrac run async_proof.nra` (`async function f(): int { return 1 }`, `const x = await f()` in `export async function run`) prints `async_proof.nra:1:1: expected a declaration, found "async"`, exit 1.
- branch: the same program prints `1`, exit 0.

## Notes

- Nothing of the node existed: no TAsync, no async keyword, only the E0930 check for a nested `async function` (which read `async` as an identifier).
- AST: `FuncDecl.isAsync`, `Expr.AwaitExpr(id, pos, operand)` and `Expr.AsyncExpr(id, pos, body)`, where body is the ArrowExpr or the BlockExpr after `async`. Wrapping keeps the arrow machinery (closure facts by arrow id, hints) and lets fmt print the source as written. The closure facts of an async block are kept by the AsyncExpr's id (`findTaskClosureId` in check.nra).
- Plain arrows need parentheses, so `async x => ...` is not parsed either.
- Dead end found while planning the tests: with D109 as written, no task in async-core can ever wait. The creation function runs the body until it first waits, and a handle exists only after that first run returns, so the first await of a program always meets a finished handle. A cycle of awaits, a task left waiting at exit and a suspended frame dropped at exit all need something that waits without a task. Decision: build `Async.idle()` in its ready-queue form (it finishes once nothing is ready), the function event-loop would have added. event-loop extends it with timers and I/O. Recorded as a deviation.
- Exit drop: phase 1 releases what every unfinished frame holds, phase 2 drops the loop's references. A task waiting on its own handle through a Cell (frame holds the Cell, Cell holds the handle) is then freed: no leak.
- Async functions skip the `from` (shared results) check: the value goes through the task, which holds its own copies of the parameters. The shared-write check treats `await` as sharing nothing and a handle as sharing every argument (calls) and capture (blocks).
- Operand spilling: calls and templates already hold earlier arguments when a later one may write (`mayWrite`, whose default arm is true for await). Added: the left operand of a binary operator and the earlier fields of a struct literal, in a task body only. Array literals store each element before the next is evaluated, so they keep their order. `&&`, `||` and `if` values were already branches.
- `emitgenerators.nra` imported `declaredLocal` from emitexprs; emitasync needs recordLive and emitStateSwitch from emitgenerators, and emitexprs dispatches to emitasync, so the lookup moved to `findDeclaredLocal` in emitstate.nra to break the cycle.
- A debug-built compiler (`narrac build` without -release) overflows the C stack in the const-eval depth tests; the suites must use a release build.
- Open doubt: idle's rule here (finish once nothing is ready) is looser than D109 item 7 (no other task unfinished). With only a ready queue the two differ only for tasks stuck in a cycle; event-loop must decide when timers and I/O exist.

## Review fixes (paused 2026-11-13)

- Done: (1) an expression statement whose type holds an Async<T> is E1103 whatever its syntax; (2) `_`, `const _ =`, `let _ =`, `_`-prefixed names and destructured `_` parts that hold an Async<T> are E1103 (`testdata/errors/async_discarded_handle`); (3) map keys are held before an awaiting value (`emitasyncspill.nra`, `emitmaps.nra`, sample `testdata/async_map_await`, README gap removed); (4) `await` and `async` in `AST_KINDS`.
- All four are in 97a04089 (squashed, rebased on feat/batch-12 d6311cc7, rebase exit 0). Branch compiler still emits byte-identical C for the base selfhost tree and the 236 non-async samples; the targeted suites pass.
- Not done: the full check (`sh scripts/test.sh`) on the rebased commit was stopped at the pause. Next step: run it in the foreground, then squash this log commit into the node commit and report.
