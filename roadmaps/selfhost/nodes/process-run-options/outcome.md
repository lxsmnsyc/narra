# Outcome: process-run-options

## Built

`Process.run` takes options for input, environment, working folder and inherited output. Calls with two arguments compile and behave as before.

## Public API

```narra fragment
export function run(
  command: string,
  arguments: Array<string>,
  input: Option<string> = None,
  environment: Array<#[string, string]> = [],
  removed: Array<string> = [],
  folder: Option<string> = None,
  inheritOutput: bool = false,
): Result<#[int, string, string], string> uses io, alloc
```

- Options are passed by name: `Process.run("git", ["status"], folder: Some(path))`, `Process.run("cat", [], input: Some(text))`.
- `environment` sets variables (a later pair of a name wins); `removed` drops names. Windows matches names without case.
- A set variable without a name, a removed name that is empty or holds `=`, and a folder that does not exist are `Err`.
- With `inheritOutput` the child writes to this process's stdout and stderr, in order (pending output is flushed before the child starts); the captured strings are empty.
- No `Process.RunOptions` struct: SPEC allows only constant defaults, and a struct literal default (`RunOptions {}`) is rejected (E1212). Flat defaulted parameters give the same call forms.

## Where

- `std/src/ext__process.nra`: the Narra part of std/process. `run` lives here now.
- `selfhost/src/std/stdprocess.nra`: the builtin is `Spawn.run` in the new module `std/internal/process`; `Process.cwd` and `Process.exit` stay builtin. The runtime `selfhost/runtime/std_process.c` loads for `std/internal/process`.
- `selfhost/runtime/std_process.c`: POSIX parent writes input and reads both streams in one `poll` loop (SIGPIPE blocked while writing; a pending one is dropped), the child gets `chdir` and a new `environ` before `execvp`. Windows: input pipe with a writer thread, environment block, working folder, standard handles passed through for `inheritOutput`.
- `selfhost/src/check/buildconditions.nra`: the Wasi E0991 rule now names the function `std/ext/process|run`; the message still reads `Process.run`. `Process.run` stays E0991 on Wasi-Wasm32.
- `testdata/process_run_options.nra` with its `.out`: 1 MiB through `cat`, env add/replace/remove, `pwd` in a folder, passed-through output order, bad inputs.
- SPEC.md (section on `std/process`), `docs/guide/11-files-and-the-system.md`, generated `docs/reference/std/process.md`, `stdtext.nra` and `runtimestdprocess.nra`.

## Deviations

- Flat parameters instead of an options struct (see above).
- Touched compiler files outside the brief's paths: `selfhost/src/std/stdprocess.nra` (declared as owned in practice), `stdlib.nra`, `emit/emitstd.nra`, `runtime/runtimetext.nra`, `check/buildconditions.nra`. Each change is a few lines.
- The Windows path (CreateProcessW with environment block, folder, input pipe and writer thread, inherited handles) compiles: the C of `testdata/windows_process_options.nra` and `testdata/process_run_options.nra` built with `narrac build -target Windows-X86_64 -emit-c` compiles and links with `zig cc -target x86_64-windows-gnu`. It was not run: no wine or Windows host here. `testdata/windows_process_options.nra` is in `scripts/cross-check.sh` for that.
- Program lookup: both systems search the PATH the child gets, after `removed` and `environment` apply. A child with no PATH keeps the default search (`execvp` on POSIX, `CreateProcessW` on Windows).
- The pending SIGPIPE: `sigtimedwait` on Linux, `F_SETNOSIGPIPE` on macOS (no `sigtimedwait` there).

## Gotchas

- The seed does not know the Narra part of std/process, so the compiler's own calls (`Process.run` with two arguments, the `sh -c` sites) are untouched. The seed update is the orchestrator's: after it, `selfhost/src` can use the options.
- Regenerating `stdtext.nra` needs a compiler that knows the change. Build one with the seed (`narrac0 emit selfhost`, `cc`), use it for `--suite generated --update`, then build again.

- `Process.run` of a program that exits without reading its input is fine; the leftover input is dropped.
