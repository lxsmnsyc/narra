# Log: process-signals

- Proof on feat/batch-9 (09baaae6), narrac built by `sh selfhost/bootstrap.sh`: spinning program (`while (true)`), `narrac run spin.nra &`, wait 10 s, `kill -TERM <narrac pid>`, check the narrac child. Result: narrac gone, child (pid 28067) still ALIVE after 2 s.
- Same script on the branch build: child GONE.
- Review finding: `narrac test` is NOT covered by the Process.run fix. The test binary forks one child per test (test_runner.c) and Test.panics forks too (std_test.c); those children were orphaned.
- Choice: forward SIGTERM, SIGINT, SIGHUP to the child from handlers installed during `narra_proc_run`; the child stays in the caller's process group, so terminal Ctrl-C still reaches it directly. Signals are blocked from before fork until the child pid is recorded; handlers and mask are restored afterwards and in the child before exec.
- Test: `exits` suite case `signal_forwarding` (selfhost/tests/src/exits.nra) runs testdata/signals/forward.sh against testdata/signals/spin_until_killed.nra.
- Test sample fixed: marker path is args[1] (args[0] is the program). exits suite passes.
- Review round: rebased on feat/batch-9. Same forwarding (block, install, fork, attach pid, wait, detach with signals blocked) added to test_runner.c and std_test.c. The runner stops after a caught signal and returns 128 plus the signal; Test.panics exits the same way. std_process.c clears the child slot before restoring handlers and mask. forward.sh now takes run|test, checks every descendant of narrac, and has a trap.
- Proof (forward.sh, all descendants of narrac must be gone 10 s after SIGTERM): base build, run: pid 50552 left; base build, test: pids 51716 51753 left (test binary child and per-test child). Branch build: run stopped, test stopped.
- Review round: NARRAC=<branch narrac> sh scripts/test.sh PASS, 7061 cases, 0 failed in every shard.
- Linux review round: forward.sh no longer uses pgrep; the samples report their own pid (sh -c 'echo $PPID'); SIG_IGN signals are left ignored; Test.panics flushes stdout before exiting. test.sh PASS 7061 cases.
