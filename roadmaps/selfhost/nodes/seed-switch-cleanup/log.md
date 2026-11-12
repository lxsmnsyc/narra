# Log: seed-switch-cleanup

- Not already fulfilled: runProgram returned 1 on three paths and buildtarget compared text.
- Before (narrac from the branch base): `narrac run -stdin /nonexistent/x prog.nra` prints `narrac: /nonexistent/x: No such file or directory` (from sh) and exits 1; startProgram itself succeeds there, so it does not reach the failed-start branch. I found no input that makes Process.run fail after a good build. The failed-start branch printed the error and returned 1 (full release).
- After: same output and exit 1, with and without NARRA_CHECK_LEAKS. The failed-start branch now calls exitWithoutRelease(1), as do the temp-directory and build-error returns in runProgram.
- buildtarget.nra: TargetRow gains hostOs and hostArch values; detectHostTarget compares them with Build.hostTarget(); describeOs and describeArch removed. NARRA_HOST_TARGET still matches the text.
