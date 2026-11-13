# Outcome: async-api

## Built

- `std/async` (`std/src/async.nra`, embedded in `stdtext.nra`), beside async-core's `AsyncError`, `detach` and `idle`:
  - `then<T, effect E>(handle: Async<T>, escaping callback: (value: T) => void uses E) uses E, alloc`. The callback's effects are charged at the `then` call. It wraps the callback in a closure over the finished handle and passes it to `Tasks.then`. For an `Async<void>` the callback takes no argument: `Async.then(work(), () => ...)`.
  - `peek<T>(handle: Async<T>): Option<T> uses alloc`. It gives `Some(value)` only when the task is done and not cancelled.
  - `cancel<T>(handle: Async<T>)`.
  - `all<T>(handles: Array<Async<T>>): Async<Array<T>> from handles uses alloc`. This is an `async { ... }` block that awaits each handle in array order.
  - `race<T>(handles: Array<Async<T>>): Async<Option<T>> from handles uses alloc`. This is an `async { ... }` block. With no handles it gives `None`. Otherwise it awaits `Tasks.winner()`, which a watcher on each handle sets to that handle's index, then gives `Some(await handles[index])`.
  - `detach`, `then` and `cancel` take an `Async<void>`. `peek`, `all` and `race` on one are E1150 (below).
  - Neither `all` nor `race` cancels its inputs. A cancelled input still gives its value, because both await.
- Primitives (`selfhost/src/stdasync.nra`, module `std/internal/async`):
  - `Tasks.then`, `Tasks.cancel`, `Tasks.finished`, `Tasks.value` and `Tasks.watch` use the check kind `checkTasks` (`checkTasksCall`), because their types depend on the handle's `T`.
  - `Tasks.winner` is a plain row giving `Async<int>`.
  - `emitAsyncCall` now takes the walker.
  - `Tasks.value` is a borrowed read, `(*(T*)(handle)->value)`, valid only after `Tasks.finished`.
- Runtime (`selfhost/runtime/std_async.c`):
  - New header fields: `callbacks` and `last_callback`, a list of `narra_task_callback { fn, env, release, next }` kept in registration order.
  - `narra_task_then(task, fn, env, release)`. On a cancelled task it drops the callback. On a finished one it queues the callback for the next loop turn, as TypeScript does: a small `narra_task_later` task in the ready queue holds the callback and a reference to the handle. Otherwise it appends the callback.
  - `narra_task_cancel(task)`. It does nothing on a done or already-cancelled task. Otherwise it sets `cancelled`, drops the callbacks without running them, then calls event-loop's `narra_loop_cancel_wait(task)`, which wakes a timer, descriptor or idle wait of the task.
  - The race winner has its own no-op release, `narra_task_winner_release`. `narra_loop_cancel_wait` treats a task whose release is `narra_task_release_nothing` as an idle handle, so a cancel must not mistake a winner for one.
  - `narra_task_finished(task)` is `done && !cancelled`.
  - `narra_task_finish` queues the waiters, then runs the callbacks unless the task is cancelled, then drops them, and only then drops the loop's reference. The list is taken off the task before it runs, so a callback may call `then` or `cancel` on the same handle.
  - Race helpers: `narra_task_winner_new()`, a bare task with an `int64_t` slot. `narra_task_watch(winner, handle, index)` sets the winner at once if `handle` is done; otherwise it makes a watcher task that waits on `handle` with `narra_task_wait`, and the watcher sets the winner when it resumes, unless another handle came first. The winner keeps a doubly linked list of its watchers. When it is settled, every other watcher still in a waiter list is unlinked from that list and from the loop and freed, so a race decided early leaves nothing on a handle that stays pending. A watcher already queued ends when it resumes.
  - `narra_task_check_watchers(live)`, called at the start of `narra_async_exit`: under `NARRA_CHECK_LEAKS`, a watcher whose race is decided but that is still live is reported as `leak check: N race watchers outlived their race`, exit 102. Exit would free such watchers, so the ordinary leak check cannot see them.
- Exit: `narra_async_exit` (`std_async_loop.c`) runs the watcher check, then marks each live task with `narra_task_mark_cancelled`, so the callbacks of tasks left at exit are dropped too. Two lines.
- `Async<void>` in generic calls (D109 ruling, 2026-11-13):
  - Inference (`generics.nra`, `unify`): a type parameter that is the value of an `Async<T>` binds to `void` when the argument is an `Async<void>`. Elsewhere void still binds nothing.
  - A function-type parameter whose type is void takes no argument, so `(value: T) => void` with `T = void` is `() => void`. Both `substitute` and `resolveType` (`resolve.nra`) drop it.
  - A call through a function value (`closures.nra`, `checkCallThroughValue`) lets a void argument fill no parameter only in a generic instance body, and only when the argument's type in the generic body is a type parameter (`isVoidParameterValue` in `asyncvoid.nra` finds the generic body with the same node ids). So the generic body `callback(Tasks.value(finished))` checks in the void instance, while `g(voidCall())` with `g: () => void` stays E0316, as in TypeScript. The emitter (`emitclosures.nra`) evaluates a void argument as a statement and passes nothing. `Tasks.value` of an `Async<void>` is `((void)0)`.
  - E1150 (`selfhost/src/asyncvoid.nra`, `checkVoidBindings`, called from `calls.nra`): a type parameter inferred as void may appear in the signature only as `Async<T>` or as a parameter of a function type. Anywhere else, such as `Option<T>` or `Array<T>`, the call is E1150, with its row and explanation in `diagnostics.nra`. The check covers only parameters that the signature names in an `Async<T>`; any other void binding comes from an error already reported, such as `third<4>(1)`.
- Samples (`.nra` and `.out`):
  - `testdata/async_api_then`: pending, finished, a callback that holds its own handle, a callback that starts a task.
  - `async_api_cancel`: callbacks and a waiter attached when the cancel comes; `peek` before the end, after the cancel and after the end; `await` on a cancelled handle; a double cancel; `then` after the cancel; cancel on a finished handle.
  - `async_api_all`: out-of-order finish, empty array, `Arrays.map` with an async arrow, a cancelled input.
  - `async_api_race`: first value, empty array, finished handles in array order, a race left waiting at exit.
  - `async_api_exit`: a self-waiting task with callbacks that hold its handle, and an `all` and a `race` over it, left at exit.
  - `async_api_void`: `then`, `detach`, `cancel`, a queued `then` and a user generic function over `Async<T>`, all on `Async<void>` handles.
  - `async_api_cancel_wait`: a cancel ends a one-hour `AsyncTime.sleep` with `Err(AsyncError.Cancelled)`, and wakes a task waiting in its own `Async.idle()` while another task keeps the loop busy. A `then` on the cancelled sleeper does not run.
  - `async_api_race_many`: 2000 races against a handle that never finishes, under the watcher check. Before the fix, with the sweep removed, it ends `leak check: 2000 race watchers outlived their race`, exit 102. With the fix it is clean.
  - `testdata/errors/async_api_then_type`: a callback of the wrong type is an ordinary type error.
  - `testdata/errors/async_api_void_peek`, `_all`, `_race`: E1150.
  - `testdata/errors/async_api_void_argument`: `g(nothing())` with `g: () => void` is E0316.
- Codes: E1150 (a type parameter inferred as void stands for a value). E1151-E1169 stay free.

## How cancel marks a task (for std-async-io)

- `Async.cancel(handle)` calls `narra_task_cancel`. It sets `task->cancelled = true`, frees the `then` callbacks, and calls `narra_loop_cancel_wait(task)`. Exit marks every task with `narra_task_mark_cancelled` (flag and callbacks, no wake) and then finishes the waits itself in timer order (`narra_loop_cancel_all_waits`), so the exit order of event-loop's samples stays the same.
- What the task sees:
  - A task parked on a timer or a descriptor wait (`narra_wait`, also behind a std function task with `forwards_cancel`) is woken: the wait finishes with -1, and `std/async/time` turns that into `Err(AsyncError.Cancelled)`. An I/O wait of std-async-io gets the same treatment if it is a `narra_wait`.
  - A task waiting in `Async.idle()` is woken: the idle handle finishes at once (`async_api_cancel_wait`).
  - A task waiting on another task's handle keeps waiting.
  - A later suspending call checks `narra_async_loop.current->cancelled` (or the started-cancelled flag) and returns `Err(AsyncError.Cancelled)` without waiting.
- `cancelled` is never cleared. `narra_loop_add` sets it on tasks created while exiting. `peek` gives `None` from the cancel on, and `then` on a cancelled handle drops its callback without running it.

## Proof

Feature node, no fix.

- base: 53a1f53a (feat/batch-12). A program calling `Async.then` fails to check: `std/async` has no `then`.
- branch: `narrac run testdata/async_api_then.nra` prints `testdata/async_api_then.out`.

## Deviations

- Files outside the brief's paths:
  - `selfhost/src/stdcalls.nra`: one `if` for `checkTasks`.
  - `selfhost/src/emitstd.nra`: the `Tasks.` dispatch passes the walker.
  - `selfhost/runtime/std_async_loop.c` and its generated `runtimestdasyncloop.nra`: exit runs `narra_task_check_watchers` and marks tasks with `narra_task_mark_cancelled`.
  - `compiler/README.md`: the async I/O line no longer lists `then`, `peek`, `cancel`, `all` and `race` as missing; event-loop's text for timers and I/O is kept.
- `all` and `race` use `async { ... }` in `std/src/async.nra`. That file is compiled only by the compiler under test, never by the committed seed, which only builds `selfhost/`. A C task cannot build an `Array<T>` or an `Option<T>` for each type. The header comment of `std/src/async.nra` says this.
- Files outside the brief's paths, for the `Async<void>` ruling: `selfhost/src/generics.nra`, `resolve.nra`, `calls.nra`, `closures.nra`, `emitclosures.nra`, and the new module `selfhost/src/asyncvoid.nra`.
- The `Async<void>` ruling listed `Async.all([voidTask()])` among the calls that must compile. It is E1150 instead, and so are `peek` and `race` on an `Async<void>`. Their results would be `Array<void>` and `Option<void>`, which are not types. Giving `all` a different result type for void, such as `Async<void>`, would mean a second signature for the same name, and Narra has no overloading or hidden implementations. The message names the function and the parameter, and the explanation says to await the handles one by one or give the task a value.

## Stubs left


## Gotchas

- `then` callbacks of a pending handle run inside `narra_task_finish`, so user code runs there. A `then` on a finished handle runs on the next loop turn instead. The task's loop reference is held until the callbacks have run, so a callback that drops the last handle does not free the frame under the loop.
- A callback closure's C type is `void (*)(narra_env*, narra_task*)` for every `T` (handles are `narra_task*`). `std_async.c` names only `struct narra_env`, because the closure runtime may come later in the C file. `release` is passed in from the call site.
- Watchers are live tasks with `resume` set (`narra_task_watcher_resume`), and each holds a reference to the winner. Settling the winner frees the watchers that lost. event-loop: `narra_async_exit` must keep calling `narra_task_check_watchers` before it cancels anything.
- A generic function whose signature uses `T` only as `Async<T>` or as a function-type parameter works for `Async<void>`. A body that names a `T` value, for example `const x: T = await handle`, is not covered in the void instance.

## Proposed learnings

- A std primitive whose type depends on a type parameter needs a `check:` kind with a case in `stdcalls.nra`, and its emitter needs the walker | StdFunc rows have no type parameters | selfhost/src | no

## Verification

- `NARRAC=<release branch narrac> sh scripts/test.sh --filter async`: PASS, 194 cases. This covers the programs suite under AddressSanitizer, the leak check and the copy check, in debug and release.
- After both review rounds, rebased on feat/batch-12 89b07514 (event-loop landed): `sh scripts/test.sh` (narrac built from the seed, bootstrap job included): PASS, 8434 cases in 97s.

## Size

- Estimated M, actual M.
