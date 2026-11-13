# Log: async-generators-design

- 2026-11-13: branch renamed to spike/async-generators-design, reset to feat/batch-13 (b23bbd4f). DECISIONS.md has no async generator decision (D95 and D109 item 4 only defer it; E1107 rejects `async function*`).
- Read: D17, D19, D31, D78, D85, D88, D93, D95, D109; SPEC sections 2 and 8; `emitgenerators.nra`, `emitasync.nra`, `asynccheck.nra`, `stmts.nra` (`checkYield`, E1105), `parser.nra` (E1107), `contracts.nra` (`seedLoopFacts`), `std_async.c`, `std_async_loop.c` (`narra_loop_cancel_wait`), `runtime.h` (`narra_gen`), `std/src/async.nra`, `std/src/coroutine.nra`.
- Key constraint found: a `Generator<T>` is a value (cloned before resume when shared, `narra_gen_next`), but a task frame cannot be cloned while it sits in a waiter list or on a timer. So `AsyncGenerator<T>` is an identity type like `Async<T>`.
- The cancel walk in `narra_loop_cancel_wait` already handles "forwarding" tasks (`forwards_cancel`) and the "only waiter" rule; a pending step fits as a forwarding task whose `awaited` is the stream.
- Measured: `sizeof(narra_task)` = 128, `sizeof(narra_gen)` = 48 (cc, macOS arm64, `/private/tmp/claude-501/async-generators-design/size.c`, struct copied from `std_async.c`).
- Counts: `Type.TGenerator` 58 matches in 22 files; `Stmt.ForStmt(` 32 matches in 22 files.
- SPEC sample written as a `text` block: a `ts` block of syntax that does not parse would need a `spec-skip.txt` line, a file outside this spike's paths. for-await turns it into `ts`.
- Split: async-generators (L, E1112-E1121) and for-await (M, E1122-E1129).
- Checks: `sh scripts/test.sh --suite spec` PASS 62 cases; `--suite sources` PASS; `--suite codes` PASS.
