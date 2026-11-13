# Outcome: event-loop

## Built

- The loop, `selfhost/runtime/std_async_loop.c` (generated into `runtimestdasyncloop.nra`, picked with the task text by `runtimeFor("std/internal/async")`). One `_Thread_local narra_loop` per thread holds the ready queue, the unfinished tasks (`live`), the idle handles, a timer heap and the poller. `kqueue` under `__APPLE__`, `epoll` under `__linux__`, `#error` elsewhere. The poller descriptor opens on the first wait and closes in `narra_async_exit`.
- A turn (`narra_async_run_entry`): run the ready queue in order until it is empty; else wait in the poller until the nearest timer deadline or a ready descriptor (`narra_loop_poll`, false when no timer and no descriptor wait exist); else finish the oldest idle handle (`narra_loop_wake_idle`); else stop with `async: every task waits on another task`, exit 101.
- Timers: a binary min-heap of `narra_wait*` ordered by deadline, then by a creation number, so equal deadlines fire in the order they were set. A deadline past the clock's range saturates at `INT64_MAX` (it waits until a cancel). The heap array is `malloc`ed and freed at exit.
- Poll: one wait in the poller lasts at most a day (`NARRA_POLL_LIMIT`), then the loop waits again. `EINTR` counts as no event; any other poller error stops the program with `async: the poller failed: <strerror>` and exit 101 instead of spinning.
- Wait handles: `struct narra_wait { narra_task task; int64_t value; started; deadline; sequence; slot; int descriptor; bool writable; next_io; previous_io; }`. A wait handle is a bare task (resume NULL, release `narra_wait_release`, value slot `&wait->value`), so Narra code sees it as `Async<int>` and awaits it like any handle. `narra_wait_is(task)` tells one apart. Its value: 0 or more on success (the nanoseconds slept; 0 for a ready descriptor), -1 when cancelled, an errno value above 0 when the poller refused the descriptor.
- C API for later nodes:
  - `narra_task* narra_async_sleep(int64_t nanos)`: timer handle (zero or less fires on the next turn, after the ready tasks).
  - `narra_task* narra_loop_wait_io(int descriptor, bool writable)`: the I/O registration for std-async-io. One reader and one writer may wait on a descriptor at once on both platforms; a second wait in the same direction finishes at once with `EEXIST` (checked in the loop's io list). kqueue: one one-shot event per direction, udata = the wait. epoll: one registration per descriptor (`data.fd`), its mask the directions still waited for (`narra_epoll_sync`: ADD, MOD or DEL); a wake or cancel removes only its own direction. The handle finishes with 0 when ready.
  - `narra_loop_cancel_wait(narra_task* task)`: how a cancelled wait wakes. Call it after setting `task->cancelled`. When the task waits in its own `Async.idle()` handle, that handle leaves the idle list and finishes (orchestrator ruling, 2026-11-13). Otherwise it follows `task->awaited` through tasks marked `forwards_cancel` down to a wait handle, takes that wait out of the heap or the poller and finishes it with -1, so the std function awaiting it resumes and returns `Err(AsyncError.Cancelled)`. A task that awaits a user task's handle keeps waiting (D109 item 8).
  - `narra_loop_cancelled_now()`: true when the running task, or the task that started it, is cancelled, or the program exits; the wait functions then return a handle already finished with -1.
- Task header (`std_async.c`): two new fields. `started_cancelled` is set in `narra_loop_add` when the task that starts the new one is cancelled; `forwards_cancel` is set by `narra_async_sleep` and `narra_loop_wait_io` on the running task, which is always the task of a suspending std function, since only `std/` modules can import `std/internal/async`. Together they make a cancel reach the wait inside `AsyncTime.sleep` without passing to user tasks (a user task started by a cancelled task keeps its own state; only std primitives read `started_cancelled`).
- `Async.idle()` (D109 item 7 as amended by the maintainer, 2026-11-13): idle handles finish one at a time, oldest first, each once nothing is ready and no timer or descriptor wait is pending. A task started while an idle handle waits (also one started by the waiter of an older idle handle) holds the later ones. In a cancelled task and at exit it finishes at once (unchanged).
- Exit (D109 items 6, 8, 10), `narra_async_exit`: mark every live task cancelled; repeat {finish every timer and descriptor wait with -1, run the ready queue} while an idle handle is left to wake; then release every unfinished frame, drop the waiter lists and the loop's references; free the heap; close the poller. Tasks sleeping at exit get `Err(AsyncError.Cancelled)`, every later sleep returns it at once, and they run to their return. The entry's value stays the exit code.
- `std/async/time` (`std/src/async__time.nra`, embedded in `stdtext.nra`, module row in `ASYNC_FUNCS`): `export async function sleep(duration: Duration): Result<Duration, AsyncError> uses io, alloc`; `Ok` holds the time slept (never less than asked). It awaits the primitive `Tasks.sleep(nanoseconds: int): Async<int>` (row in `INTERNAL_ASYNC_FUNCS`, effects `io, alloc`, emitted as `narra_async_sleep(...)`). The module uses async syntax: std sources under `std/async/` are compiled only by stage 1, never by the seed.
- `std/src/async.nra`: the `idle` comment states the new rule. `emitAsyncCall` takes the walker now (`emitstd.nra` passes it).
- Samples: `testdata/async_time_longest` (a sleep of `INT64_MAX` nanoseconds saturates and is cancelled at exit; before the fix it woke with `Ok` after microseconds), `testdata/async_time_order` (deadline order, equal durations in set order, zero sleeps), `async_idle_started` (idle waits for sleeping tasks and tasks started while it waits, oldest idle first), `async_exit_sleeping` (sleepers cancelled at exit, a second sleep cancelled at once, a heap string held across the sleep released, an hour-long sleep and a forever ticker do not keep the program running); `testdata/exits/async_exit_sleepers_code` (code 3 while cancelled tasks return other values), `testdata/exits/async_loop_cycle_after_timer` (a pending timer runs before the cycle stop, exit 101).
- `compiler/README.md`: the "Not supported yet" async I/O line describes the new loop; the async-core idle difference line is gone (the exit-release sentence stays as its own line).
- D109 item 7 (`DECISIONS.md`, with the status line "Amended 2026-11-13 (maintainer): idle rule.") and the SPEC section 8 `Async.idle` and cancel text state the accepted idle rule.
- `roadmaps/selfhost/nodes/event-loop/io_wait_test.c` and `io_wait_test.sh`: a C-level test of `narra_loop_wait_io` over a socketpair, under AddressSanitizer and UBSan: a reader waits, a second reader gets `EEXIST`, a writer waits beside the reader and wakes alone, the reader wakes on a byte, a cancelled wait gives -1 and leaves the poller, the direction can wait again, `narra_async_sleep(INT64_MAX)` saturates, and exit cancels it. Not part of `scripts/test.sh`.

## Deviations

- Idle rule: D109 item 7 said idle finishes when "no other task of the loop is unfinished". Taken literally, two tasks in `Async.idle()` at once wait on each other forever, and a task in a cycle of awaits never finishes. Built instead: the loop must be quiescent (nothing ready, no timer, no descriptor wait), and idle handles finish one at a time, oldest first. The maintainer accepted it; D109 and SPEC section 8 are amended.
- The cancel of a task waiting in `Async.idle()` (`narra_loop_cancel_idle`) and cancel through `Async.cancel` are not tested: async-api builds `cancel`. Cancel is tested through the exit sequence, which uses the same wait-cancel path (`narra_loop_cancel_all_waits` instead of `narra_loop_cancel_wait`).
- Files outside the brief's paths: `selfhost/src/emitstd.nra` (passes the walker to `emitAsyncCall`), `compiler/README.md`, `DECISIONS.md` (D109 item 7 and status) and `SPEC.md` (section 8 idle and cancel text), at the orchestrator's request, `testdata/exits/async_exit_sleepers_code.*` and `testdata/exits/async_loop_cycle_after_timer.*` (exit-code cases live in `exits/`), and the goldens `testdata/errors/unknown_module.err` and `testdata/declsdump/imports/err_unknown_module.dump` (the std module list now ends with `std/async/time`).
- No codes used from E1130-E1149.
- `scripts/linux-check.sh` unchanged.

## Stubs left

- No std module registers a descriptor yet: `narra_loop_wait_io` is compiled into every async program but only std-async-io calls it. The descriptor paths run only in `io_wait_test.c` (passes on macOS and in the Linux image).
- `Async.cancel` must call `narra_loop_cancel_wait(task)` after setting `cancelled` (async-api, or whichever of the two lands second).

## Gotchas

- async-api: `Async.cancel(handle)` = set `handle->cancelled`, drop the `then` callbacks, then `narra_loop_cancel_wait(handle)` (wakes an idle, timer or descriptor wait). Do not finish wait or idle handles yourself: they may be in the heap, the poller or the idle list. Orchestrator ruling: async-api changes `std_async_loop.c` too (exit cancels through `narra_task_cancel`, so `then` callbacks of tasks stuck at exit are freed); the second to land rebases and keeps both, so `narra_task_cancel` should call `narra_loop_cancel_wait`, and the exit loop here (`narra_loop_cancel_all_waits` then run ready) stays.
- std-async-io: write each suspending function as an `async` source function that awaits `Tasks.*` primitives returning `Async<int>` from `narra_loop_wait_io`, then does the nonblocking read, write, accept or connect itself; map -1 to `Err(AsyncError.Cancelled)` and a positive value to `Err(AsyncError.Failed(strerror))`. Set the descriptor nonblocking before waiting. Check `narra_loop_cancelled_now()` semantics: a primitive called from a cancelled task returns a handle already finished with -1.
- `forwards_cancel` is set on whatever task is running when a wait primitive is called, so call them only from the task of the std async function itself, never from a helper task.
- `Async.detach(f())` with `f(): Async<void>` fails with "cannot tell type parameter T of detach here"; the samples use `int` results. Not fixed here.
- A timer of zero still waits for the ready tasks to run first, so a zero sleep is a yield.
- epoll waits in whole milliseconds, rounded up; kqueue in nanoseconds. Samples assert order, never durations, except that `Ok` is at least the duration asked (holds on both).

## Proposed learnings

- Runtime waits the loop owns are bare tasks with an `int64_t` value | a std async function awaits them as `Async<int>` and builds its own `Result`, so the C runtime never needs the C names of Narra enum types | selfhost/runtime | no

## Verification

- `NARRAC=<release narrac built from the branch> sh scripts/test.sh`: PASS, 8361 cases.
- `sh scripts/linux-check.sh` (colima, native arm64 Linux, epoll path): `test.sh: PASS, 8366 cases in 245s`, `linux-check: PASS in 543s`.
- `sh roadmaps/selfhost/nodes/event-loop/io_wait_test.sh`: `io_wait_test: PASS` on macOS and in the `narra-linux-check` image.
- `sh scripts/test.sh` (narrac from the seed, macOS arm64, bootstrap job included): `test.sh: PASS, 8366 cases in 252s`.

## Size

- Estimated M, actual M.
