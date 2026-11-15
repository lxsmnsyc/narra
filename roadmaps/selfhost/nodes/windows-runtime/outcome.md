# Outcome: windows-runtime

## Built

- `selfhost/runtime/runtime.h`: on Windows, a constructor sets stdin, stdout and stderr to binary mode, and the default file mode is binary.
- `selfhost/runtime/std_fs.c`: `narra_mkdir` calls `mkdir` with one argument on Windows; text and byte files open in binary mode.
- `selfhost/runtime/std_env.c`: arguments from `__argv` on Windows.
- `selfhost/runtime/std_entropy.c`: `rand_s` on Windows.
- `selfhost/runtime/std_path.c`: `\` is a separator and a leading drive such as `C:` is kept in front of the result on Windows; other systems unchanged.
- `selfhost/runtime/std_test.c` and the generated `selfhost/src/runtime*.nra` files updated (`--suite generated --update`).
- Samples: `testdata/windows_runtime.nra` (files, env, clock, entropy, threads) and `testdata/windows_path.nra` (path rules; each line checks the rule of the system it runs on, so one `.out` serves both).
- SPEC section 3, bullet "std/path and Windows".

## Proof

Not a fix node.

## Acceptance

- Samples give the same output on Windows, run by cross-targets' script: not met here. Evidence so far: both samples build with `zig cc -target x86_64-windows-gnu` (PE32+ x86-64) and run natively with output equal to the goldens. The run on Windows needs `scripts/cross-check.sh` and `.github/workflows/windows.yml`, which node cross-targets owns; the `windows-latest` job there is the proof. wine was tried twice in docker and failed (wrong binary name, then a docker disk error), so no run of the `.exe` has happened.
- `std/path` on Windows in SPEC section 3 with samples: met.
- `sh scripts/test.sh` on the host: met after the format fix. Run on the rebased branch: every shard passed except `sources/formatted` (`testdata/windows_path.nra`), fixed with `narrac fmt`. After the fix: `--suite sources` PASS, 1 cases; `--suite generated` PASS, 23 cases; `--suite guide` PASS, 330 cases.

## Review

A fresh agent found no must-fix item. Fixed the should-fix items: `dirname` of a bare drive, UNC and stdio binary mode in SPEC. Not run on Windows: no Windows-only code path has executed, only compiled.

## Deviations

- No workflow file written: the brief's CI proof is the cross-targets job, so adding one here would duplicate it.
- `targetCompilerArguments` stays empty for Windows: the samples link with no extra flags.

## Stubs left

- Windows processes and async I/O: node windows-process-async.
- The first real run on Windows is the cross-targets CI job.

## Gotchas

- Runtime text edits need two seed updates at the end of the batch (batch-16 brief).

## Proposed learnings

- None.

## Verification

- `sh scripts/test.sh` on the rebased branch: all 8 shards 1236 or 1237 passed; one failure `sources/formatted`, since fixed.
- `sh scripts/test.sh --suite generated`: PASS, 23 cases. `--suite guide`: PASS, 330 cases. `--suite sources`: PASS, 1 cases.
- Final full run: all shards passed except a regex timing check (`(x+x+)+y` under 1 s) under load 50; `--filter regex` PASS, 63 cases.

## Landing

Landed on feat/batch-16 by the orchestrator on 2026-11-15. Acceptance item 1 (same output on Windows) moves to the windows-latest job of cross-targets, which first runs when the batch pull request opens. A mismatch there reopens this node. Under wine in docker, `scripts/cross-check.sh` ran both samples and every other sample of the list on Windows-X86_64 with matching output (10 of 10, seed check skipped), so the runtime itself is shown to work.
