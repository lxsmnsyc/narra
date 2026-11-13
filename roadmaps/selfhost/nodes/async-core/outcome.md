# Outcome: async-core

## Built

- Syntax: `async` and `await` are keywords (`lexer.nra`, `SYMBOL_KEYWORD_ASYNC`, `SYMBOL_KEYWORD_AWAIT`). Top-level `async function` and `export async function` (`FuncDecl.isAsync`), `async (x) => ...` and `async { ... }` (`Expr.AsyncExpr(id, pos, body)`, body is the ArrowExpr or BlockExpr), and `await operand` at unary precedence (`Expr.AwaitExpr(id, pos, operand)`). Cases in the checker, emitter, fmt (`async function`, `await x`, `async ...`), astdump (`func ... async`, `await`, `async` nodes), idcheck, typesdump and every other Expr walker.
- Type: `Type.TAsync(value: Type)` (`types.nra`, key `Y(...)`), written `Async<T>` and resolved as a builtin name in `resolve.nra` (a user type named `Async` hides it); `Async<void>` is a task with no value. C type `narra_task*` (`cnames.nra`, id `Y_<value>`). No automatic Equatable or Hashable (`traits.nra`).
- Checks: `selfhost/src/asynccheck.nra` (`checkAwait`, `reportUnusedHandle`, `holdsHandle`, `checkDiscardedName`, `checkDiscardedPattern`); async signature (`resolve.nra`: result `Async<T>`, E1104, E1111); async body (`body.nra`: returns checked against `T`, `Checker.asyncBody`); async arrows and blocks are checked as arrows (`closures.nra`, `checkAsyncExpression`; `ClosureInfo.isTask`); E1110 in `funcvalues.nra` (`useLocal`, `markChangedCapture`); E1103 in `stmts.nra` for any expression statement whose type holds an `Async<T>` (also inside a tuple, array, Option, struct field or map), and for `_`, `const _ =`, `let _ =` or a `_`-prefixed name (also in a destructuring) that holds one; E1105 in `stmts.nra`; E1106 in `traits.nra`; an async entry in `scopes.nra` (`checkEntry`). Effects (`effects.nra`): a call of an async function charges its effects plus `alloc`; an async block charges its body plus `alloc` where it is written; an async arrow's set gets `alloc`. Shared writes (`sharedwrites.nra`): a call whose result is `Async<T>` shares every argument, an async block shares its captures, `await` shares nothing.
- Emitter: `selfhost/src/emitasync.nra` (`emitAsyncFunction`, `emitTaskBlock`, `emitTaskArrow` returning `TaskCall { creation, inputs }`, `emitAwait`, `emitTaskReturn`). Shared with generators: `recordLive` and `emitStateSwitch` are exported from `emitgenerators.nra`; `Emitter.inGenerator` is renamed `inFrame`, and `Emitter.inAsync` and `taskValue` mark a task body (saved in `BodyContext`). Names: an async function `f_H_name` has frame `T_H_name`, resume `r_H_name`, release `q_H_name`; the n-th async arrow or block in a function has frame `U..._n`, resume `u..._n`, release `p..._n` and creation `h..._n`. Spilling in a task body: the left operand of a binary operator, earlier struct literal fields and a map key before an awaiting value are held in frame temporaries and consumed after the await (`selfhost/src/emitasyncspill.nra`: `containsAwait`, `holdBeforeAwait`, `holdArgument`, used by `emitexprs.nra` and `emitmaps.nra`); calls and templates hold through `mayWrite`.
- Entry (`emit.nra`, `emitMain`): an async entry runs `narra_async_run_entry(run(args))`; a program whose checker used `std/internal/async` calls `narra_async_exit()` after the entry, before `narra_consts_release`, the copy report and the leak report. Programs with no async emit the same C (see Verification).
- Runtime, picked by `runtimeFor("std/internal/async")` (task text, then loop text), included only when the emitter notes the helper `std/internal/async`:
  - `selfhost/runtime/std_async.c` (`runtimestdasync.nra`): `struct narra_task { rc; state; done; cancelled; dropped; resume(narra_task*); release(narra_task*); value; awaited; next_waiter; waiters; next_ready; queued; next_live; previous_live; }`. Functions: `narra_task_init(task, resume, release, value)` (rc 2: the handle and the loop), `narra_task_start(task)` (runs the body until it first waits), `narra_task_wait(task, handle)` (retains the waiter), `narra_task_finish(task)` (waiters to the ready queue in wait order, drops the loop's reference; nothing may read the task after it), `narra_task_retain`, `narra_task_release`. A frame is `{ narra_task h; T value; fields... }`; `h.value` points at `value` (NULL for `Async<void>`); release releases the value once `done`, else what is live at `state`.
  - `selfhost/runtime/std_async_loop.c` (`runtimestdasyncloop.nra`): `static _Thread_local narra_loop narra_async_loop` with `ready_head/ready_tail`, `live` (unfinished tasks), `idle_head/idle_tail`, `current`, `exiting`. API: `narra_loop_add`, `narra_loop_remove`, `narra_loop_queue(task)`, `narra_loop_enter(task)`/`narra_loop_leave(previous)`, `narra_loop_run_ready()`, `narra_loop_wake_idle()`, `narra_async_idle()`, `narra_async_run_entry(entry)` (deadlock stop: `async: every task waits on another task` on stderr, exit 101), `narra_async_exit()` (cancel every live task, run ready and idle until neither is left, release every unfinished frame, clear waiter lists, then drop the loop's references).
- `std/async` (`std/src/async.nra`, embedded in `stdtext.nra`): `enum AsyncError { Cancelled, Failed(message: string) }`, `detach<T>(_handle: Async<T>)` (an empty function: passing the handle uses it), `idle(): Async<void> uses alloc`. Primitives: builtin module `std/internal/async` with the row `Tasks.idle` (`selfhost/src/stdasync.nra`: `ASYNC_FUNCS`, `INTERNAL_ASYNC_FUNCS`, `emitAsyncCall`), registered in `stdlib.nra`; the emit dispatch is the `Tasks.` case in `emitstd.nra`. async-api and event-loop add their primitives as `Tasks.*` rows in `stdasync.nra` and functions in `std/src/async.nra`.
- Codes (`diagnostics.nra`, rows and explanations): E1100, E1102, E1103, E1104, E1105, E1106 (two formats: `==` and map key), E1107, E1108, E1109, E1110 (two formats: capture, change), E1111 (new: a function-type parameter of an async function must be `escaping`). E1101 stays free; E1112-E1129 are free.
- Samples: `testdata/async_functions`, `async_arrows`, `async_blocks`, `async_detach`, `async_plain_entry`, `async_exit_drop`, `async_nested_await`, `async_map_await` (`.nra` and `.out`); `testdata/exits/async_exit_code` (code 5) and `async_await_cycle` (code 101); `testdata/errors/async_*` for E1100 (three places), E1102 to E1111, discarded handles (`async_discarded_handle`), E0978 and the effects; `testdata/astdump/async_syntax`, `testdata/fmt/async_syntax`.
- `compiler/README.md`: "Not supported yet" drops async functions and async arrows (threads and channels stay, async I/O and the rest of `std/async` are listed with their nodes); the D85 line names E1105; a new "Differences" line covers `Async.idle` and the exit drop.

## Proof

- base: d9d1bf12 (feat/batch-12), `narrac run async_proof.nra` with `async function f(): int { return 1 }` and `const x = await f()` in `export async function run(_args: Array<string>): int`: `async_proof.nra:1:1: expected a declaration, found "async"`, exit 1.
- branch: the same command prints `1`, exit 0.

## Deviations

- `Async.idle()` is built here, though D109 gives it to event-loop. Without it no task can wait in async-core (a handle exists only after its task's first run has returned, so every await meets a finished handle), and the acceptance cases for a cycle, a task left waiting and a frame dropped mid-expression cannot be written. Here it finishes once no task is ready, at once in a cancelled task and at exit; D109 item 7 says "no other task of the loop is unfinished". event-loop owns the final rule.
- New code E1111 (a non-`escaping` function parameter of an async function), from this node's free block; the generator rule E0615 names generators in its message.
- Files edited outside the brief's paths: `selfhost/src/body.nra` (async body result, `asyncBody`, no `from` check), `scopes.nra` (async entry), `emitconsts.nra` (async function dispatch), `emitstd.nra` (`Tasks.` dispatch), `stdtypes.nra` (`Async<...>` row type), and mechanical Expr arms in `trivia.nra`, `audit.nra`, `consteval.nra`, `constfunctions.nra`, `fmtprinter.nra`, `inlining.nra`, `operands.nra`, `querysites.nra`. Test files outside `testdata/async_*`: `testdata/exits/async_*`, `testdata/astdump/async_syntax.*`, `testdata/fmt/async_syntax.*`, and the goldens `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` (the std module list now ends with `std/async`).
- Async functions skip the shared-results (`from`) check: the value passes through the task, which holds its own copies of the parameters.
- After the review, `selfhost/src/emitmaps.nra` and `selfhost/tests/src/dumps.nra` were added to this node's paths: map keys are held before an awaiting value, and `AST_KINDS` lists `await` and `async`.
- E1103 covers every discarded value whose type holds an `Async<T>` (orchestrator ruling), not only a bare handle.

## Stubs left

- No I/O, no timers, no `Async.then`/`peek`/`cancel`/`all`/`race`: event-loop and async-api. `cancelled` is set at exit but only `idle` reads it.
- The seed is not regenerated; the seed rule holds (no async in `selfhost/src` or `std/src`).

## Gotchas

- event-loop: `narra_async_run_entry` and `narra_async_exit` call `narra_loop_run_ready` and `narra_loop_wake_idle`; the poller wait goes between them, and the deadlock stop must also check timers and registered I/O. Exit phase 1 calls `release` on every unfinished frame, including ones waiting on a timer: remove them from the poller first. Bare idle tasks have `resume == NULL`; never queue one.
- async-api: add header fields for `then` callbacks to `struct narra_task` and run them in `narra_task_finish` before the loop's reference is dropped. `detach` is an empty function; `cancel` must also wake a task waiting in idle.
- A handle's value is read through `*(T*)handle->value`; the awaiting task retains it. A frame is never copied (no clone).
- In a task body every C variable is a frame field (`declareVariable`), and `recordLive` panics if a counted value held at an await is not one. A new expression form that evaluates an operand into C text and uses it after evaluating a later operand must hold the earlier one when the later holds an `await` (`containsAwait`, `holdArgument`). Array literals need nothing: each element is stored before the next runs.
- `noteAsyncUse` (check.nra) adds `std/internal/async` to `checker.used` whenever a type `Async<T>` is resolved, an await is checked or an async arrow or block is checked; main calls `narra_async_exit()` only then, so plain programs keep their C.
- A debug-built compiler overflows the C stack on the const-eval depth tests; use `narrac build -release` for `NARRAC=`.

## Proposed learnings

- An `Expr` variant added to the AST needs an arm in every walker, also those with a `_` arm | `_` arms silently skip the new node in id collection, effects and shared writes | selfhost/src | no (the compiler only reports matches without `_`)

## Verification

- Branch compiler (release, built from the branch source) emitting the feat/batch-12 `selfhost` tree: byte-identical to the base compiler's output (216309 lines, `cmp` clean). The same holds for all 236 top-level `testdata/*.nra` samples that do not use async.
- `NARRAC=<branch narrac> sh scripts/test.sh`: PASS, 8320 cases (every suite, including generators, coroutines, sources, codes, fmt, dumps, exits, errors, programs).
- `sh scripts/test.sh` (narrac built from `seed/narrac.c`, bootstrap job included): PASS, 8325 cases in 107s.

## Size

- Estimated L, actual L.
