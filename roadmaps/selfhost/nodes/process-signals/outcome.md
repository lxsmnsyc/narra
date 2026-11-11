# Outcome: process-signals

## Built

- `narra_proc_run` in `selfhost/runtime/std_process.c` (so `Process.run`, which `narrac run` and `narrac test` use) forwards SIGTERM, SIGINT and SIGHUP to the child while it waits. The child stays in the caller's process group, so Ctrl-C in a terminal still reaches it directly. The signals are blocked from before `fork` until the child pid is stored, then unblocked; handlers and the mask are restored after `waitpid`, on the start-failure path, and in the child before `exec`.
- Generated text `selfhost/src/runtimestdprocess.nra` regenerated (`--suite generated --update`). The seed is not regenerated.
- SPEC.md (`std/process` bullet): one sentence on the forwarding.
- Test: case `signal_forwarding` in the `exits` suite (`selfhost/tests/src/exits.nra`), with `testdata/signals/forward.sh` and `testdata/signals/spin_until_killed.nra`. It starts `narrac run` on a spinning program, waits until the program wrote its marker file, sends SIGTERM to narrac and requires the program to be gone; it kills the program itself if not.

## Review round

- `selfhost/runtime/test_runner.c` and `selfhost/runtime/std_test.c` (Test.panics) use the same forwarding, so `narrac test` leaves no orphan. The runner stops at once after a caught signal and returns 128 plus the signal; Test.panics ends the test the same way.
- `std_process.c` clears the child slot with signals blocked, then restores handlers, then the mask. SPEC says the caller keeps running and `run` returns 128 plus the signal.
- Test: `exits/signal_forwarding_run` and `exits/signal_forwarding_test`; `testdata/signals/forward.sh run|test` checks every descendant of narrac is gone and kills them on its own failure or kill. Sample `spin_test.nra`.
- Proof (base build vs branch build, all descendants must be gone): run base left pid 50552, test base left 2 pids; branch build stopped both.

## Proof

- base: feat/batch-9 at 09baaae6, narrac built by `sh selfhost/bootstrap.sh`. A looping program run with `narrac run spin.nra &`, 10 s wait, `kill -TERM <narrac pid>`: narrac died, its child (pid 28067) was still alive 2 s later.
- branch: same script with the branch build: the child was gone (pid 25255).
- `narrac test` has its own fork in test_runner.c; see Review round.

## Deviations

- Touched `selfhost/runtime/test_runner.c`, `selfhost/runtime/std_test.c` and their generated text for `narrac test`.
- Touched `selfhost/tests/src/exits.nra` and added `testdata/signals/` (not in the brief's paths) for the test; the exits suite is the existing runner for process-ending cases.
- `sh scripts/linux-check.sh` was not run (no container host in this loaded session). The code uses only POSIX `sigaction`, `sigprocmask`, `kill`; `forward.sh` uses `pgrep -P`, which exists on both macOS and Linux.

## Stubs left

- None. The seed needs its normal batch-end regeneration to carry the runtime change into the seed-built compiler.

## Gotchas

- The fast check with a seed-built narrac cannot pass `signal_forwarding` until the seed is regenerated, because the seed's runtime text lacks the forwarding. Stage 1 of `selfhost/bootstrap.sh` (and the `NARRAC=` build used here) has it.
- The seed-built compiler is not needed to pass the new cases once the stage 2 compiler is used by test.sh.
- The handler state is process-global: two threads calling `Process.run` at once share one child slot, and the later one wins for forwarding.
- `forward.sh` waits up to 5 minutes for the program to start, because compiles are slow on a loaded machine.

## Proposed learnings

- A test that starts a never-ending program must kill it itself on every failure path | an orphan spins for hours | selfhost/tests | no

## Verification

- `NARRAC=<branch narrac> sh scripts/test.sh` (first run): every suite passed except the new `exits/signal_forwarding`, whose sample read its marker path from `args[0]`, which is the program name (the path is `args[1]`).
- After the fix, `NARRAC=<branch narrac> sh scripts/test.sh --suite exits`: PASS (both cases), no spinner left behind. The other 8 shards of the first run were 878 to 879 passed, 0 failed. The full run was not repeated after the one-line sample fix because of the machine load.

## Size

- Estimated M, actual M.

Landed on feat/batch-9 as 754c3280. Linux check passed on the final branch (worker run, 2026-11-12).
