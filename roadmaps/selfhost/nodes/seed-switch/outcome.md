# Outcome: seed-switch

## Built

- Seed regenerated for batch-9 (54141070); one update reached "seed is current".
- `selfhost/src/emitter.nra`: `skipTeardown` and the `Retained` cell cycle are gone. `exitWithoutRelease(code): int` calls `Process.exit(code)`, or returns the code when `NARRA_CHECK_LEAKS` is set so the leak check still sees a full release. `runEmit` ends every path after `emitProgram` with it.
- `selfhost/src/driver.nra`: `CompiledProgram` carries `package` and `checker`; `buildExecutable` returns it; `runBuild` (all binaries kept in `builds`) and `runProgram` end with `exitWithoutRelease`, so build, run and test also exit with the checker live.
- `selfhost/src/buildtarget.nra`: `detectHostTarget` uses `Build.hostTarget()` instead of `/usr/bin/uname -s -m`. `NARRA_HOST_TARGET` still overrides it. `None` sets `hostError` to "no supported target for this C compiler". `TARGETS` lost its `system` and `machine` columns.
- Seed regenerated after the switch (a75ea370).
- Commits on `chore/seed-switch`: 54141070 seed, 7c0c4503 code, a75ea370 seed.

## Proof

Instructions retired by `narrac emit selfhost` (`/usr/bin/time -l`, median of 5, narrac from stage 2 C or the seed with `-O2 -DNARRA_RELEASE`):

- base: 5ab6adb8 compiler on its own source: 2.414G (2.402G to 2.435G).
- branch: a75ea370 compiler on its own source: 2.396G (2.396G to 2.423G), -0.7%.
- Same input (final source): base compiler 2.427G, branch compiler 2.396G, -1.3%.
- The branch compiler with `NARRA_CHECK_LEAKS=1` (full release): 2.450G, so the full release costs about 0.05G more than the exit.

## Deviations

- Emitter release in `emitProgram` not skipped. It is safe (keeping the `Emitter` and `ReachableBodies` in `EmitResult` only holds them until the exit; the leak check still releases them), but measured on the same source it saved 4M instructions (2.382G to 2.377G, 0.2%), within run-to-run spread, for two new `EmitResult` fields and a third `CompiledProgram` field. `emit.nra` is unchanged.
- The unsupported-host error for `None` reads "could not determine the host: no supported target for this C compiler": `reportUnsupportedHost` (`buildconditions.nra`, not in this node's paths) prefixes `hostError`.
- Edited `SPEC.md` (host-is-the-target bullet) and `compiler/README.md` (std/build entry) to drop the uname wording; neither is in the brief's paths.

## Stubs left

- None.

## Gotchas

- The host is now the target narrac was compiled for, not the machine it runs on. A narrac built for x86_64 and run under Rosetta targets MacOS-X86_64.
- A command that keeps a value for the exit must hold it in a local live at the `exitWithoutRelease` call; a value released earlier (a callee's local) is not saved.
- `exitWithoutRelease` ends the process; code after a driver command returns is never run unless the leak check is on.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh` after each seed update: "ok: stage 1 and stage 2 emit identical C", "seed is current".
- `sh scripts/test.sh`: PASS, 7107 cases in 151s.
- `sh scripts/linux-check.sh`: PASS in 316s (its bootstrap also reports the seed current).

## Size

- Estimated S, actual S.

Landed on feat/batch-9 as four commits, 54141070 to 14fbe1fb. Review notes: the exit saves 0.05G against the 2.427G same-source baseline here; `log.md` measures 0.07G against a different binary (2.382G). Two optional nits left for later: `driver.nra` returns 1 with a full release when `startProgram` fails after a build, and `buildtarget.nra` compares the host target by text.
