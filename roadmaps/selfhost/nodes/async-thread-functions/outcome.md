# Outcome: async-thread-functions

## Built

- `Thread.startAsync(scope, work)` of `std/thread` runs an `async` thread function: an `async` arrow written at the call or the name of a top-level `async` function, of this module or an imported one. Signature: `export function startAsync<T, effect E>(threadScope: Scope, escaping work: () => Async<T> uses E): Thread<T> uses E, alloc`. The thread runs the task on its own event loop until it finishes, then cancels and drains what is left, as before, and `Thread.join` gives the awaited value. An `Async<void>` task gives a `Thread<void>`.
- Checker (`selfhost/src/check/threadcheck.nra`): the thread checks run on `Thread.startAsync` too: captures must be `Sendable`, and so must the value the task gives (the `Async` result is unwrapped only for `startAsync`). `Thread.start` takes only plain functions: an `async` arrow or the name of an `async` function is E1192, `Thread.start takes a plain function; use Thread.startAsync to run an async function on a thread`, with a new explanation. A first version (27a8c706) checked `Thread.start` with async work as a call of a private `startTask`; the orchestrator rejected it (2026-11-20), since one call name then ran a different function by the kind of its argument (no hidden implementations, no overloading). `asyncvoid.nra` no longer counts `Thread<T>` as a void value, so E1150 does not fire for an `Async<void>` task. - std: `startAsync` in `std/src/thread.nra` runs `CellSlot.store(result, Some(Tasks.runUntilDone(work())))` on the new thread. `Tasks.runUntilDone` is a new primitive of `std/internal/async` (`selfhost/src/std/stdasync.nra`).
- Runtime: `narra_loop_run_until(task)` in `selfhost/runtime/std_async_loop.c` runs the loop of the thread until the task is done and gives its value; `narra_async_run_entry` now calls it. A cycle of awaits stops the program with the same message.
- `Parallel.map` stays plain. It splits work that computes into one part per core; an async callback would run the items of a part one after another on that part's loop, so waiting work gains nothing over `Async.all` on one loop. An async callback is E1190 (`Async<int> is not Sendable`). `Shared.update` and `Shared.read` callbacks stay plain (E1410).
- fmt: an `async` arrow as the last argument hugs the call, as a plain arrow does and as oxfmt does (`selfhost/src/tools/fmtexprs.nra`).
- Cases: `testdata/threads_async.nra` (sleep, loopback `std/net` echo, an async arrow that leaves a sleeping task for the thread's exit, an `Async<void>` task), `testdata/packages/threads_async_module` (an async function of another module runs on a thread), `testdata/errors/threads_async_unsendable.nra`, `testdata/errors/threads_async_function.nra` and `testdata/packages/errors/threads_async_module` (E1192 for an async arrow, an async function of this module and one of an imported module), `testdata/errors/parallel_async_callback.nra`, `testdata/fmt/async_arrow_hug`. `testdata/channel_full_blocks.nra`: the receiver is an async thread function that awaits `Time.sleep`; the busy wait is gone.
- Docs: SPEC section 8 ("Thread functions", "Tasks on a thread"), D125 item 6 and an amended line in D111, `compiler/README.md`, guide chapter 15 (the thread function rules and "Threads or async tasks", with a sample) and one sentence in chapter 14. Regenerated: `docs/reference/errors.md`, `docs/reference/std/thread.md` (`startAsync`), `selfhost/src/std/stdtext.nra`, `selfhost/src/runtime/runtimestdasyncloop.nra`.
- Cancelled await: not reproduced. SPEC section 8 says `await` on a cancelled handle gives a `T`, because the body sees `Err(AsyncError.Cancelled)` and returns. A cancel then `await` returned at once for a sleeping task, a chain of awaits, `Async.idle`, an `async` block, a bare `Time.sleep` handle, `Net.accept`, `for await` over a stream and `Async.next`. Two specified cases wait: `Async.all` and `Async.race` handles (a cancel stops there, so the await waits for the inputs, as `testdata/async_api_all_cancel` shows), and a handle that another task also awaits. A body that ignores the error and loops without a suspending call also never returns. A cancelled race over a long sleep looks like a hang but follows SPEC. `testdata/async_api_cancel` and `async_api_cancel_wait` already cover the returning cases, so no case was added.

## Proof

- base: 6752f01c, `Thread.start(scope, async () => 1)` and `Thread.start(scope, Jobs.job)` gave E1192 (`testdata/errors/threads_async_function`, `testdata/packages/errors/threads_async_module`).
- branch: `Thread.start(scope, async () => 1)` and `Thread.start(scope, Jobs.job)` give the new E1192 naming `Thread.startAsync`; `testdata/threads_async` (with `Thread.startAsync`) prints `slept 10`, `echo ping 4`, `arrow 5`, `quiet` under the sanitizer and leak check; the package prints `job 1`.

## Deviations

- fmt change (async arrow hugging), not in the brief: without it the formatter put the async arrows of the new cases on their own lines, unlike plain arrows and oxfmt.
- `Thread.startAsync` is a second public function next to `Thread.start`, by the orchestrator decision of 2026-11-20, in place of a hidden rewrite of `Thread.start`.

## Stubs left

- None.

## Gotchas

- E1192 is by syntax: a plain arrow that returns a handle (`Thread.start(scope, () => nap())`) is still a plain thread function and its `Async` result is E1190.
- `Tasks.runUntilDone` is a blocking wait. Only `std/thread` calls it, at the top of a fresh thread, where no other task can be waiting.
- The value of the task is read from its frame before the statement's handle temporary is released, as `Tasks.value` does.

## Proposed learnings

- none

## Verification

- After the startAsync change, targeted with a stage 1 compiler: errors 912, fmt 95, codes 1, generated 26, sources 1, spec 65, guide 620, packages 74, query 18 (builtin_signatures included), programs `--filter thread` 36 and `--filter channel` 16: all PASS.
- `pnpm --dir website build`: exit 0.
- Full `sh scripts/test.sh` (from the seed): `test.sh: PASS, 10898 cases in 376s`; the bootstrap fixed point holds (stage 2 and stage 3 emit identical C).

## Size

- Estimated M, actual M.
