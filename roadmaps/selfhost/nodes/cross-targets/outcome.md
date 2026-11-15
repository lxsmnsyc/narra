# Outcome: cross-targets

## Built

- `scripts/cross-check.sh [row ...]`: builds `hello arrays closures enums maps int_to_string float_shortest c_interop` for every row of `narrac targets` with `narrac build --target` and `zig cc` (a one-line wrapper in `CC`), runs each where a runner exists and compares stdout with the `.out` golden. It also builds `seed/narrac.c` for the row and, where it runs, compares the C it emits for `testdata/c_interop.nra` with the host seed's.
- Runners: host (native), Rosetta, docker `--platform linux/amd64|arm64` (alpine for musl, debian for gnu), wine in a `linux/amd64` docker image the script builds once (`narra-cross-wine`), natively on a Windows host, `wasmtime -W exceptions=y`. A missing runner is a skip that says why.
- Exit 0 only when every build and run matched its golden. Rows named in `expected_failures` (Windows-X86_64, Wasi-Wasm32) report as expected failures until their runtime nodes land and remove them; `STRICT=1` counts them as failures. `left_out` holds `row:sample` pairs for wasi-runtime. `SKIP_SEED=1` skips the seed builds.
- `.github/workflows/windows.yml`: `windows-latest`, pull requests to `main`, runs `sh scripts/cross-check.sh Windows-X86_64` with zig 0.17.1. Checked with `actionlint` (clean). Not run.
- `compiler/README.md`: says when to run the script. Not added to `scripts/test.sh`.

## Results on this machine (macOS arm64, zig 0.17.1, docker via colima, wasmtime)

Host compiler: stage 3 of `sh selfhost/bootstrap.sh` on this branch, rebased on the landed windows-runtime. Samples: `hello arrays closures enums maps int_to_string float_shortest c_interop windows_runtime windows_path` (10). Runs were split by group of rows because of the load (above 100 at times). Before the rebase the first full run took 12 minutes.

| row | build | run | note |
| --- | --- | --- | --- |
| MacOS-Arm64 | 10/10 | 10/10 ok, native | seed emits the same C as the host (checked before the rebase) |
| MacOS-X86_64 | built | none | no Rosetta installed (`arch -x86_64` fails: Bad CPU type); with REQUIRE_RUN=1 the row fails |
| Linux-Arm64 | 10/10 | 10/10 ok, docker | seed emits the same C (before the rebase) |
| Linux-Arm64-musl | built | 8/8 ok before the rebase | seed emits the same C |
| Linux-X86_64 | built | 8/8 ok before the rebase (qemu) | seed emits the same C |
| Linux-X86_64-musl | 10/10 | 10/10 ok, docker (qemu) | seed emits the same C (before the rebase) |
| Windows-X86_64 | 10/10 | 10/10 ok, wine in docker | seed build fails: the committed seed still calls `mkdir(path, 0777)`; it needs the batch seed update |
| Wasi-Wasm32 | 0/8 | none | `zig cc` rejects `setjmp.h` without `-mllvm -wasm-enable-sjlj`; seed fails the same way; expected failure until wasi-runtime |

Rows marked "before the rebase" were not rerun with the two windows samples; the other rows ran them with `SKIP_SEED=1`. The final run exits 0 (Wasi-Wasm32 is the one expected failure).

## Deviations

- Wasi-Wasm32 does not pass here: this base has no wasi-runtime. It is in `expected_failures`. `targetCompilerArguments` has no wasm flags yet (`-mllvm -wasm-enable-sjlj`, `-mexception-handling`, `-lsetjmp`); the seed needs them to build too.
- Windows-X86_64: the host and the ten samples pass under wine. Before windows-runtime landed, stdout had `\r\n` line ends and the row failed; it passes now. The seed row on Windows fails to build until the seed update at the end of the batch, so `windows.yml` (which builds narrac from the seed on the runner) stays red until then. It has not run.
- MacOS-X86_64 builds but does not run: Rosetta is not installed here.
- The wine image sets `WINEARCH=win64` and runs `wineboot --init` (the first image printed a wine prefix error). Docker runs set `TMPDIR=/tmp` because the repository is mounted read-only and `windows_runtime` writes files.
- Docker was stopped; `colima start --cpu 2 --memory 3` was run for the Linux and wine rows. It is still running.
- Overlap with windows-runtime: it wrote no workflow; `windows.yml` here is the one job. `REQUIRE_RUN=1` makes a row with no runner fail, so the job cannot pass without running.

## Review

A fresh agent (pr-review) found three must-fix and four should-fix items. Fixed: the reference seed uses `$CC`; the Windows wrapper is a `.cmd` file (untested on Windows); a row that ran nothing is marked, and fails under `REQUIRE_RUN=1`; unused variables; the README sentence on docker. Not changed: the CI job cannot fail on an expected failure now that Windows is not in `expected_failures`.

## Stubs left

- `expected_failures` still lists Wasi-Wasm32; wasi-runtime removes it. `left_out` is empty; wasi-runtime fills it.

## Gotchas

- The script is slow under load: the seed builds with `-O2` per row and Linux x86_64 runs under qemu.
- `docker run` needs the repository under the home folder (colima mounts it).

## Proposed learnings

- None.

## Verification

- `sh scripts/cross-check.sh`, `... Linux-Arm64 Linux-Arm64-musl Linux-X86_64 Linux-X86_64-musl`, `... Windows-X86_64 Wasi-Wasm32`, all with `NARRAC` set: exit 0, as in the table.
- `sh scripts/test.sh` on the rebased branch: `test.sh: PASS, 9897 cases in 264s`. Before the rebase one run failed `programs/regex_linear.nra/debug` (a 1 s timing check) at load above 150; the rerun passed.
