# Outcome: wasi-runtime

## Built

- `targetCompilerArguments` (`selfhost/src/buildtarget.nra`) gives the Wasi row `-mexception-handling -mllvm -wasm-enable-sjlj -lsetjmp`. It also has `-mllvm -wasm-use-legacy-eh=false`. `runtime.h` defines the `__c_longjmp` tag under `__wasi__` (D122), so debug builds stay at `-O0`; before that, `-O0` failed to link with undefined `__c_longjmp`.
- E0991 on Wasi for what WASI preview 1 cannot back, in `selfhost/src/buildconditions.nra` (tables `WASI_LEFT_OUT_MODULES`, `_FUNCTIONS`, `_BUILTINS`): `Process.run`, all of `std/thread`, `std/parallel`, `std/channel`, `std/shared`, `std/atomic`, `std/async/net`, and `Terminal.columns`. A use inside `if (Build.os != Os.Wasi)` is accepted, as are uses inside the left-out modules. `@target` itself could not do this: it names one OS, and `Process.run` is a builtin with no Narra source. Hook: `checkBuiltinTargetUse` in `exprs.nra`, `checkTargetUse` for Narra functions. One new E0991 row in `diagnostics.nra`.
- Runtime: `std_env.c` reads `Env.args` with `args_sizes_get`/`args_get` on wasi (bug 1); `std_fs.c` sends `/dev/stdout` and `/dev/stderr` writes to descriptors 1 and 2 (bug 2); `std_process.c` compiles on wasi with a `narra_proc_run` that returns an error; `std_terminal.c` has no ioctl on wasi (width 0); `std_async_loop.c` has no poller on wasi.
- Async loop decision: timers only, no `poll_oneoff`. A turn with nothing ready calls `nanosleep` until the nearest timer. Descriptor waits do not exist because sockets are left out and file reads are done in the call. Recorded in SPEC section 12.
- `narrac test --target Wasi-Wasm32`: not supported. `test` already rejects `--target` on every row (target-table), so no new error.
- Tests: `testdata/wasi_args.nra`, `wasi_stdout.nra` (run under wasmtime and on the host, with goldens), `wasi_cases/left_out.nra` + `.err` (`emit` case `wasi/left-out`, `check --all-targets`; outside `testdata/*.nra` so the programs suite does not build it).
- SPEC section 12 and `compiler/README.md` updated. Generated `runtime*.nra` rewritten with `--suite generated --update`.

## Proof

- base: bug 1 and 2 reproduced before the fix: the wasm program printed no `Env.args` and nothing from `Fs.writeText("/dev/stdout", ...)`.
- branch: `wasmtime run -W exceptions=y wasi_args.wasm a b` prints `run: 2`, `env: 2`, `arg a a`, `arg b b`. `wasi_stdout.wasm` prints `to stdout` and `to stderr`.

## Deviations

- Work item 1 asked for `@target` marks on the std functions. `@target` names one OS and `Process.run` is a builtin with no Narra source, so the left-out set is three tables in `buildconditions.nra`. A new std function that WASI cannot back needs a row there. A later node could let `@target` take a negated OS.
- `chooseNamedTarget` (`driver.nra`) gives a Wasi target the real target even when the program does not import `std/build`, and `check --all-targets` adds the Wasi row for such a program. Without this the E0991 checks ran only for programs that imported `std/build`, so `Atomic.create` built and panicked at run time.

- `scripts/cross-check.sh`: the Wasi row left `expected_failures` and `left_out` has `c_interop` and `windows_runtime` (coordinator request). `SKIP_SEED=1 sh scripts/cross-check.sh Wasi-Wasm32`: 8 ran, 2 left out, 0 failed. The seed build for that row waits for the seed update at the end of the batch.
- Edited outside owned paths: `selfhost/src/buildconditions.nra`, `exprs.nra`, `diagnostics.nra`, `buildtarget.nra`, `driver.nra` (E0991 for builtins and the Wasi flags; `buildtarget.nra`, `driver.nra` and the std/runtime files are in `shared`; `buildconditions.nra`, `exprs.nra`, `diagnostics.nra` are not), `selfhost/tests/src/emitchecks.nra` (one case). No new code number: the existing E0991 gets a second message form.
- The `targets` contract lists neither the extra flags nor the tag; no change request is needed, this node carries them (the brief says so).
- `-lsetjmp` is in `targetCompilerArguments` and so comes before the source file; it links fine.

## Stubs left

- `narra_proc_run` on wasi returns an error (unreachable when E0991 holds).
- Runtime text changed: the seed needs the two seed updates at the end of the batch.

## Gotchas

- Samples that touch files need `--dir .` (repo root as cwd, since some read `testdata/...`) and `--dir /tmp`.
- `Build.os` matches in samples written before Wasi existed fall to their last branch (`build_target.nra` prints `\`).
- wasm32 `size_t` is 32 bits: a sample that declares `memset` with a 64-bit size fails to run.

## Proposed learnings

- Runtime code must not read `/proc` or `/dev` paths without a wasi branch | seen again here | selfhost/runtime | no

## Verification

- Wasm sweep at the default `-O0`, `narrac build --target Wasi-Wasm32` with `zig cc` (zig 0.17.1) and `wasmtime run -W exceptions=y --dir . --dir /tmp` (50.0.1), repo root as cwd, every `testdata/*.nra` with a `.out`: 297 built, ran and matched the golden (295 plus `io_stdin` and `io_stdin_all`, which need their `.in` as stdin). 51 left out: 45 are E0991 (process, thread, parallel, channel, shared, atomic, async/net, terminal use); `coroutine_tests`, `std_test` and `tests` are `test` block files with no entry; `build_target` (its `Os` branch has no Wasi case), `c_interop` (calls libc `strlen` through a bitcast) and `unsafe_fixed_arrays` (64-bit `memset` size on wasm32) build but differ.
- `wasi_args` matches its golden (`run: 0`, `env: 0`) with no arguments; with `a b` it prints `run: 2`, `env: 2`, `arg a a`, `arg b b`. `wasi_stdout` matches its golden.
- `sh scripts/test.sh` on the rebased branch: one run failed `programs/regex_linear/debug`, a timing case ("under 1 s") on a machine at load 20 to 30. Before the rebase and review fixes it printed `test.sh: PASS, 9905 cases in 308s`. After them, `--suite programs` printed `test.sh: PASS, 713 cases in 165s`, `--suite generated` `PASS, 23 cases` and `--suite guide` `PASS, 330 cases`. Shards of the failed run: every other case passed.
- Debug-built compiler (`narrac build selfhost`), `NARRA_CHECK_COPIES=1 narrac-debug emit selfhost`: exit 0.
- `check --target Wasi-Wasm32` on a `Process.run` sample: E0991 (`wasi/left-out`, `wasi/left-out-plain` cases).

## Size

- Estimated M, actual M.
