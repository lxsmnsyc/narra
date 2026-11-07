# Log: linux-check

- colima profile `default` (4 CPUs, 5 GiB) was stopped; started it unchanged.
- Tests call `cc` (or `$CC`) with `-std=c11 -w` and, in the debug mode, `-fsanitize=address,undefined`. Image: `golang:1.23-bookworm` plus gcc. Linux gcc has LeakSanitizer, so the leak check runs for real.
- First full run (gcc, arm64): 33 minutes, three kinds of failure:
  1. `CLOCK_MONOTONIC` undeclared in the `narra test` runner C (`test_runner.go`). `-std=c11` hides POSIX names on glibc. Fix: `#define _POSIX_C_SOURCE 200809L` at the top of `runtime.h`.
  2. Release builds of `Str.split` segfaulted at -O2. The emitter declared the temp as `lstr***` while `narra_str_split` returns `lstr**`, so the later `parts[i]` read had the wrong type and gcc reordered it before the stores in the inlined split (strict aliasing). Fix: temp type `lstr**` in `std_string.go` and in `selfhost/src/emitstd.nra`.
  3. `TestSelfhostRuntimeText`, `TestSelfhostEmit`, `TestSelfhostBootstrap` failed after change 1 because `selfhost/src/runtimetext.nra` is generated from `runtime.h`. Regenerated with `NARRA_UPDATE=1`.
- The stdin `fgetc` loop from PR #91 compiled and passed on Linux without changes.
- No leaks found by LeakSanitizer.
- colima shares only the home directory with containers, so ad hoc scratch mounts must live under it.
