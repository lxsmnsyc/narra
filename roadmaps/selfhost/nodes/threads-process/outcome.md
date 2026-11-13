# Outcome: threads-process

## Built

- `Process.run` (`selfhost/runtime/std_process.c`), the test runner (`test_runner.c`) and `Test.panics` (`std_test.c`) forward SIGTERM, SIGINT and SIGHUP safely for calls from several threads. Each file keeps a mutex-guarded count of active calls (the first call saves and installs the handlers, the last call restores them), a fixed array of 64 child-pid slots claimed with an atomic compare-and-swap, and `pthread_sigmask` in place of `sigprocmask`. The handler walks the slots with atomic loads and `kill`; it takes no lock. A call that finds all 64 slots taken yields until one frees.
- A signal that arrives while a slot is claimed but the pid is not stored yet (it can land on another thread) is kept in a per-slot pending cell and sent by `attach`.
- The child side of the fork restores the saved handlers and the mask without taking the lock.
- Generated files rewritten: `selfhost/src/runtimestdprocess.nra`, `runtimestdtest.nra`, `runtimetestrunner.nra`.
- No new link flag: `pthread.h` and `stdatomic.h` link on macOS with no flag. The Linux result is in the final report.

## Proof

- base: 7df2cd63. `/private/tmp/claude-501/threads-process/two.c` (not committed) includes `runtime.h` and the base `std_process.c`, runs `narra_proc_run` from two pthreads (each child traps SIGTERM and exits 7) and sends SIGTERM to the process once both children run. Built with `cc -std=c11 -g -fsanitize=address,undefined -pthread`. Result: one child never gets the signal, the program hangs until its 15 s alarm, exit 142.
- branch: same program, 5 runs: both children report code 7, the SIGTERM action is SIG_DFL again, the SIGTERM mask is clear, exit 0, clean under the sanitizers.

## Deviations

- None. The pending-signal cell is an addition the brief did not name: without it, a signal delivered to another thread between `fork` and the pid store is lost (the old code avoided this only because one thread blocked the signals).

## Stubs left

- None.

## Gotchas

- A signal that reaches another thread while the last call restores the handlers runs the old handler with no child left, so it is dropped instead of ending the process. The window is the few instructions of `detach`. Node threads can close it if it matters (re-raise from `detach` when a signal was seen).
- `narra_runner_caught` and `narra_panics_caught` stay single plain variables: the runner is one caller, and `Test.panics` reads its flag only after its own call.
- Runtime text changed, so the batch needs its two seed updates at the end.
- A Narra sample for concurrent `Process.run` belongs to node threads, once there is a thread API.

## Proposed learnings

- A handler that forwards to children walks a fixed atomic array and a per-slot pending cell, never a lock | a signal can land on any thread between fork and the pid store | selfhost/runtime | no

## Verification

- Emit comparison: `narrac emit` of `testdata/arrays.nra`, `arg_eval_order.nra` and `async_api_race.nra` is byte-identical between the base and branch compilers. `narrac emit selfhost` differs only in the three runtime blocks changed here (127 changed lines, all in `narra_proc_*`, `narra_runner_*`, `narra_panics_*` and their includes).
- `NARRAC=<branch release narrac> sh scripts/test.sh --suite generated`: PASS. `--suite exits` (the signal tests: SIGTERM, SIGINT, SIGHUP for `Process.run`, `narrac run`, `narrac test`): PASS, 7 cases.
- Full check and Linux check: see the final report.

## Size

- Estimated S, actual S.
