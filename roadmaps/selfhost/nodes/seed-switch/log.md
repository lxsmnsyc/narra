# Log: seed-switch

## Instructions, method

`bench/run.sh` measures the compiler row with `/usr/bin/time -l narrac emit selfhost`. Instructions here are the "instructions retired" line of `/usr/bin/time -l`, median of 5 runs, with narrac built from the stage 2 C (`cc -O2 -DNARRA_RELEASE -std=c11 -w`). Script: `/private/tmp/claude-501/seed-switch/measure.sh <narrac> <selfhost>`.

## Before (feat/batch-9 at 5ab6adb8, no change)

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/seed-switch/before`, then cc on its `stage2.c`.
- Load average 8.16 (1 minute) at the start.
- Runs: 2.402G, 2.407G, 2.414G, 2.416G, 2.435G. Median 2.414G.

## First seed update

- `sh scripts/seed-update.sh`: stage 1 C differed from stage 2 (old seed), stage 2 and stage 3 identical (187,486 lines), seed written.
- `sh selfhost/bootstrap.sh`: "ok: stage 1 and stage 2 emit identical C (187486 lines)", "seed is current". One update was enough: the runtime text changes were already in the stage 1 compiler's output after one pass.
- Committed as 54141070 `seed: regenerate for batch-9`.

## Code change

- `emitter.nra`: `skipTeardown` and `Retained` removed; `exitWithoutRelease(code)` calls `Process.exit(code)` unless `NARRA_CHECK_LEAKS` is set, in which case it returns the code and the command releases everything.
- `runEmit` returns `exitWithoutRelease(...)` on every path after `emitProgram`, so the checker and the package are live at the exit.
- `driver.nra`: `CompiledProgram` carries `package` and `checker`; `buildExecutable` returns the `CompiledProgram`; `runBuild` keeps each binary's build in `builds` and `runProgram` keeps `built` until it returns `exitWithoutRelease(...)`.
- `buildtarget.nra`: `Build.hostTarget()` in place of `/usr/bin/uname -s -m`; `NARRA_HOST_TARGET` still wins. The `system` and `machine` columns of `TARGETS` and `UNAME_PATH` are gone. None gives `hostError` "no supported target for this C compiler" (printed as "could not determine the host: no supported target for this C compiler" by `reportUnsupportedHost`).
- `SPEC.md` and `compiler/README.md`: the host-is-the-target bullet no longer mentions uname.

## Emitter release in emitProgram

Tried: added `emitter: Emitter` and `reachable: ReachableBodies` fields to `EmitResult`, so the callers hold them until the exit. Safe (the values are only held longer; under the leak check they are released when the caller returns). Measured on the same source:

- without: 2.381G, 2.381G, 2.382G, 2.382G, 2.391G (median 2.382G)
- with: 2.377G, 2.377G, 2.377G, 2.378G, 2.411G (median 2.377G)

0.2% (4M instructions), below the run-to-run spread seen in the before runs. Not kept: two more fields on `EmitResult` and a third on `CompiledProgram` for 0.2%.

For comparison, the same binary with `NARRA_CHECK_LEAKS=1` (full release): median 2.450G, so the exit saves about 0.07G (2.8%) of release work.

## Second seed update

- Code committed as 7c0c4503.
- `sh scripts/seed-update.sh`: stage 1 and stage 2 identical (187,528 lines), seed written. `sh selfhost/bootstrap.sh`: identical, "seed is current".
- Committed as a75ea370 `seed: regenerate after seed-switch`.

## After

- narrac built from the new seed, `narrac emit selfhost` on the final source, load average 5.48: 2.396G, 2.396G, 2.396G, 2.397G, 2.423G. Median 2.396G.
- Same input, the before compiler (5ab6adb8): median 2.427G. Same-input change: -1.3%.

## Checks

- `sh scripts/test.sh`: PASS, 7107 cases in 151s.
- `sh scripts/linux-check.sh`: PASS in 316s; its bootstrap printed "seed is current", so the `std/build` import adds no host-dependent C.
