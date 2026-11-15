# Outcome: windows-process-async

## Built

- `selfhost/runtime/std_process.c`: `Process.run` on Windows. `CreateProcessW` with a command line quoted so `CommandLineToArgvW` gives the arguments back, stdin from `NUL`, one pipe per output stream with a thread that drains it, a handle list so only those three handles pass to the child, and a job object that ends the child with this process.
- `selfhost/runtime/test_runner.c`: `narrac test` on Windows starts the binary again for each test (`NARRA_TEST_CHILD` is the case number, `NARRA_TEST_PIPE` an inherited pipe handle for the panic message). A job object ends a test, and what it started, on a timeout. Option parsing moved into `narra_test_parse`, shared with the POSIX runner.
- `selfhost/runtime/std_test.c`: `Test.panics` on Windows runs the closure in the same process under a jump buffer, the one `Co.resume` uses. `narra_test_fork` is a macro there.
- `selfhost/runtime/std_async_loop.c`: the loop waits with `WSAPoll` over the io list, one entry per wait, and sleeps when only timers are set.
- `selfhost/runtime/std_async_io.c`: Winsock calls, WSA errors mapped to errno values, and Winsock started once. `SO_REUSEADDR` is not set on Windows.
- `selfhost/src/buildtarget.nra` and `driver.nra`: `targetLinkArguments` gives `-lws2_32` for Windows; the driver appends it after the C file, for a named Windows target and for a Windows host.
- Generated `selfhost/src/runtime*.nra` rewritten (`--suite generated --update`).
- `scripts/cross-check.sh`: `windows_process` and `windows_async` added to the sample list.
- Samples `testdata/windows_process.nra` and `testdata/windows_async.nra` with goldens.
- SPEC section 3, bullet "`std/process`, `std/test` and `std/async` on Windows" (known gaps); `compiler/README.md` points to it.

## Proof

Not a fix node.

## Acceptance

- `seed/narrac.c` built for Windows compiles the samples and the output matches the host's: met for the samples. `SKIP_SEED=1 sh scripts/cross-check.sh Windows-X86_64` under wine: `Windows-X86_64  0 built, 12 ran, 0 skipped, 0 failed` and `cross-check: PASS (0 expected failure rows)`. NARRAC was a compiler built from this branch's sources, so it carries the new runtime text. The seed itself was not built for Windows (SKIP_SEED), because the committed seed still has the old runtime; that check waits for the batch seed update.
- The process and async samples pass on Windows: met under wine (above). Not run on a real Windows host; the `windows-latest` job of the batch pull request is the first. Known gaps are in SPEC.
- `sh scripts/test.sh` passes on the host: met. `NARRAC=<compiler built from this branch> sh scripts/test.sh`: `PASS, 9907 cases in 320s`. `--suite guide`: `PASS, 330 cases`. `--suite generated`: `PASS, 23 cases`. `--suite sources`: `PASS, 1 cases`.
- Extra checks under wine, not in the sample list: a test binary with 8 tests (pass, failed assertion, `Test.panics`, `Test.panicsWith`, wrong message, plain panic, a 5 s sleep with `NARRA_TEST_TIMEOUT=2`) prints the same lines as the host. A child program printed eight hard arguments (spaces, quotes, trailing backslashes, empty, tab) unchanged.

## Review

A fresh agent reviewed the diff. Must-fix, fixed: nested `Test.panics` left `narra_test_current` on the inner record, so the outer return restored the wrong jump buffer; each record now keeps the previous one. A nested case runs the same under wine as on the host (`nested panics` in the ad hoc runner file). Should-fix, fixed: `SO_ERROR` mapping no longer maps a failed `getsockopt` twice; the 2 GiB length clamp and `recv` apply only on Windows (POSIX keeps `read` and `send` with `size_t`); a failed drain thread falls back to draining inline; the test runner sets its environment variables only after the checks that can fail. Added to SPEC: `.bat` and `.cmd` go through cmd.exe, a refused connect needs Windows 10 2004 or later, and the job object also ends processes the child left running. Left as is: the test child inherits all inheritable handles (the runner starts one test at a time); a failed `realloc` while draining truncates output; `setjmp` in the `Test.panics` macro is outside the ISO C forms but works with clang and gcc.

## Deviations

- The brief names the target-table link arguments: `-lws2_32` is added through `targetLinkArguments` in `buildtarget.nra` and the driver, as the coordinator decided, not through a pragma.
- `Test.panics` does not use a child process on Windows: a closure cannot be re-run in a new process. It runs in place with a jump, so it has no timeout and keeps changes to globals (SPEC lists this). The brief's "re-run the binary with the test name" applies to the test runner, which does it.
- `WSAPoll` is enough; no IOCP node is split off.
- Process signals are not forwarded on Windows: there is no SIGTERM or SIGHUP; a console passes Ctrl-C to the child, and the job object ends the child when this process ends.
- The process sample builds its scratch path from `TEMP`, or from `TMPDIR` under wine, which has no `TEMP`. A `TEMP` path with spaces would break `cmd /c type`.

## Stubs left

- The first run on real Windows is the `windows-latest` job; a mismatch reopens this node.
- The seed build for Windows (`cross-check.sh` without `SKIP_SEED`) after the batch seed update.

## Gotchas

- Runtime text edits need the batch seed updates at the end of the batch (batch-16 brief).
- `getenv` returns memory that `_putenv` can free; read the value before clearing it.
- wine has no `sort.exe` and no `TEMP`, and its `cmd` does not run `for /L ... & ...` correctly.
- `Env.args()` and the `args` of `run` hold the program path as item 0 on every system.

## Proposed learnings

- None.

## Verification

See Acceptance.
