# Log: windows-process-async

- Runtime C: `_WIN32` branches in std_process.c, test_runner.c, std_test.c, std_async_loop.c, std_async_io.c. Generated `runtime*.nra` rewritten with `--suite generated --update`.
- Build flag: `targetLinkArguments` in buildtarget.nra gives `-lws2_32` for Windows; the driver appends it after the C file for a named Windows target and for a Windows host (coordinator decision; a `#pragma comment` was tried first and dropped).
- wine in docker: the image `narra-cross-wine` exists. /private/tmp is not shared with colima, so ad hoc programs go in on stdin as a tar.
- Found under wine: wine has no `sort.exe` and no TEMP, and its `cmd` fails `for /L ... & ...`. The process sample uses `type`, `set /p` and a scratch path that allow for these.
- Bug found by the runner check: `getenv` memory is invalid after `_putenv`, so the child read its handle after clearing the variable and exited 2. Fixed by reading both values first.
- Runner check (ad hoc, 8 tests: pass, assertion fail, panics, wrong message, plain panic, timeout): output under wine equals the native output line for line, including the 2 s timeout.
- Argument quoting check (ad hoc): the child printed `a b`, `c"d`, `e\`, `f\\g`, empty, `h\"i`, a tab and `x y\` unchanged.
- Env.args() and run(args) include the program path as item 0 on every system.
