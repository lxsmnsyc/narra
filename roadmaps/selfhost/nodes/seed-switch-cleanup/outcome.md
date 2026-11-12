# Outcome: seed-switch-cleanup

## Built

- `selfhost/src/driver.nra`: every `return 1` in `runProgram` (temp directory, build error, failed start) is `exitWithoutRelease(1)`. Under NARRA_CHECK_LEAKS it still returns for the full release.
- `selfhost/src/buildtarget.nra`: `TargetRow` has `hostOs: Os` and `hostArch: Arch`; `detectHostTarget` compares them with `Build.hostTarget()`. `describeOs` and `describeArch` are gone. The `NARRA_HOST_TARGET` override still names a row by text.

## Proof

Refactor, no failing test. `narrac run -stdin /nonexistent/x prog.nra` behaves the same before and after (message from sh, exit 1, with and without NARRA_CHECK_LEAKS). A real startProgram failure after a good build could not be provoked, so that branch is covered by reading only.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- `runBuild` has early `return 1` lines (load errors, usage) that still release fully; the brief named only the run path.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh`: fixed point reached; seed not regenerated.
- `narrac fmt` on both files: no change.
- `sh scripts/test.sh`: PASS, 7107 cases.

## Size

- Estimated S, actual S.
