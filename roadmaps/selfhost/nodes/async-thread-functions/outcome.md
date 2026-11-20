# Outcome: async-thread-functions

## Built

- A thread function given to `Thread.start` may be `async`: an `async` arrow written at the call or the name of a top-level `async` function, of this module or an imported one. The thread runs the task on its own event loop until it finishes, then cancels and drains what is left, as before, and `Thread.join` gives the awaited value. An `Async<void>` task gives a `Thread<void>`.
- Checker: `findTaskStart` in `selfhost/src/check/threadcheck.nra` sees the async work before the arguments are checked, and `checkFunctionCall` (`selfhost/src/check/calls.nra`) checks the call as one of `startTask`, a private function of `std/thread`. The thread checks run on `startTask` too: captures must be `Sendable`, and so must the value the task gives (the `Async` result is unwrapped only for `startTask`). E1192 is removed: its rows in `selfhost/src/driver/diagnostics.nra` and its tests. `asyncvoid.nra` no longer counts `Thread<T>` as a void value, so E1150 does not fire for an `Async<void>` task. `unused.nra` does not report `startTask` (E0947), since only the compiler calls it.
- std: `startTask` in `std/src/thread.nra` runs `CellSlot.store(result, Some(Tasks.runUntilDone(work())))` on the new thread. `Tasks.runUntilDone` is a new primitive of `std/internal/async` (`selfhost/src/std/stdasync.nra`).
- Runtime: `narra_loop_run_until(task)` in `selfhost/runtime/std_async_loop.c` runs the loop of the thread until the task is done and gives its value; `narra_async_run_entry` now calls it. A cycle of awaits stops the program with the same message.
- `Parallel.map` stays plain. It splits work that computes into one part per core; an async callback would run the items of a part one after another on that part's loop, so waiting work gains nothing over `Async.all` on one loop. An async callback is E1190 (`Async<int> is not Sendable`). `Shared.update` and `Shared.read` callbacks stay plain (E1410).
- fmt: an `async` arrow as the last argument hugs the call, as a plain arrow does and as oxfmt does (`selfhost/src/tools/fmtexprs.nra`).
- Cases: `testdata/threads_async.nra` (sleep, loopback `std/net` echo, an async arrow that leaves a sleeping task for the thread's exit, an `Async<void>` task), `testdata/packages/threads_async_module` (moved from the errors folder: an async function of another module runs on a thread), `testdata/errors/threads_async_unsendable.nra`, `testdata/errors/parallel_async_callback.nra`, `testdata/fmt/async_arrow_hug`. `testdata/channel_full_blocks.nra`: the receiver is an async thread function that awaits `Time.sleep`; the busy wait is gone. Removed `testdata/errors/threads_async_function`.
- Docs: SPEC section 8 ("Thread functions", "Tasks on a thread"), D125 item 6 and an amended line in D111, `compiler/README.md`, guide chapter 15 (the thread function rules and "Threads or async tasks", with a sample) and one sentence in chapter 14. Regenerated: `docs/reference/errors.md`, `selfhost/src/std/stdtext.nra`, `selfhost/src/runtime/runtimestdasyncloop.nra`. The std reference did not change (`startTask` is private).
- Cancelled await: not reproduced. SPEC section 8 says `await` on a cancelled handle gives a `T`, because the body sees `Err(AsyncError.Cancelled)` and returns. A cancel then `await` returned at once for a sleeping task, a chain of awaits, `Async.idle`, an `async` block, a bare `Time.sleep` handle, `Net.accept`, `for await` over a stream and `Async.next`. Two specified cases wait: `Async.all` and `Async.race` handles (a cancel stops there, so the await waits for the inputs, as `testdata/async_api_all_cancel` shows), and a handle that another task also awaits. A body that ignores the error and loops without a suspending call also never returns. A cancelled race over a long sleep looks like a hang but follows SPEC. `testdata/async_api_cancel` and `async_api_cancel_wait` already cover the returning cases, so no case was added.

## Proof

- base: 6752f01c, `Thread.start(scope, async () => 1)` and `Thread.start(scope, Jobs.job)` gave E1192 (`testdata/errors/threads_async_function`, `testdata/packages/errors/threads_async_module`).
- branch: `testdata/threads_async` prints `slept 10`, `echo ping 4`, `arrow 5`, `quiet` under the sanitizer and leak check; the package prints `job 1`.

## Deviations

- fmt change (async arrow hugging), not in the brief: without it the formatter put the async arrows of the new cases on their own lines, unlike plain arrows and oxfmt.
- `unused.nra` has a named exemption for `startTask`. Exporting it would put a second form of `Thread.start` in the API.

## Stubs left

- None.

## Gotchas

- The redirect is by syntax: a plain arrow that returns a handle (`() => nap()`) is still a plain thread function and its `Async` result is E1190.
- `Tasks.runUntilDone` is a blocking wait. Only `std/thread` calls it, at the top of a fresh thread, where no other task can be waiting.
- The value of the task is read from its frame before the statement's handle temporary is released, as `Tasks.value` does.

## Proposed learnings

- none

## Verification

- Targeted, with a stage 1 compiler: programs `--filter thread` 36 and `--filter channel` 16, errors 907, fmt 95, codes 1, generated 26, sources 1, spec 65, guide 612, packages 74: all PASS.
- `pnpm --dir website build`: exit 0.
- Full `sh scripts/test.sh` (from the seed): `test.sh: PASS, 10836 cases in 566s`; the bootstrap fixed point holds.

## Size

- Estimated M, actual M.
