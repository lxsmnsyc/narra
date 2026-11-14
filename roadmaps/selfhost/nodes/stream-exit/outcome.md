# Outcome: stream-exit

## Built

- At exit the wait a task is suspended in returns `Err(AsyncError.Cancelled)` as before; the task runs to its next suspending call (`sleep`, descriptor wait, `Async.idle`), and that call (`narra_loop_gate` in `selfhost/runtime/std_async_loop.c`) gives a parked wait (`narra_wait.parked`): no timer, no poller entry, skipped by `narra_loop_cancel_wait`. The task suspends for good and the existing exit drop releases its frame without resuming.
- `narra_loop_cancelled_now()` is true only outside exit (a cancelled task) or for a plain function at exit, so the `Handles.cancelled` checks of std/async/file and net do not return Err at exit: file calls run their blocking work, net calls go on to their wait and park.
- Sample `testdata/async_exit_ignored_cancel.nra` and `.out`: a detached task that ignores its Err, a reader of a stream whose body ignores it, and a task that checks it and prints its cleanup.
- D109 item 6 amendment (2026-11-14) in `DECISIONS.md` and the exit bullet of SPEC.md section 8. Regenerated `selfhost/src/runtimestdasync.nra` and `runtimestdasyncloop.nra`.

## Proof

- base: c4884fe1, `narrac build testdata/async_exit_ignored_cancel.nra`, then run the binary: still running after 8 s (killed).
- branch: same command: exits, prints `keeps looping`, `entry returns`, `checker: cleanup runs`, `cleanup starts` (the await after it drops the task); clean under AddressSanitizer and the leak check in the `programs` suite (debug and release).

## Deviations

- None. The runtime text changed, so the batch needs its two seed updates; the seed was not regenerated here.

## Stubs left

- None.

## Gotchas

- A task that loops without ever calling a suspending function is a plain infinite loop.
- A parked wait is a live task in the loop's list until the exit drop; `parked` must stay set or a cancel walk would finish it.
- Goldens changed by the rule: `async_exit_sleeping`, `async_io_exit` and `async_api_race` (its "never awaited finishes" line) lost the lines printed after a second suspending call at exit (the task is dropped there; in `async_api_race`, `after` idles in a loop).
- `testdata/async_exit_sleeping` lost its "cancelled again at once" lines: the second sleep now drops the task.

## Proposed learnings

- A suspending std call runs as a task of its own, so per-caller state is read from the outermost starter | found while counting cleanup turns | selfhost/runtime | no

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh --filter async`: PASS, 549 cases.
- `narrac emit selfhost` from the base and branch compilers: byte-identical.
- Full check and Linux check: see the final report.

## Size

- Estimated S, actual S.
