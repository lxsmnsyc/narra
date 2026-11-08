# Outcome: test-runner-core

## Built

A test runner written in Narra, `selfhost/tests/` (package `testrunner`, binary `testrunner`, entry `main.run`), and `scripts/test.sh`. It runs programs, panics, compile errors, packages, `runner/` and one sanitized self-build, with narrac and no Go call. 513 cases pass.

Files in `selfhost/tests/src/`:

| Module | Role |
|---|---|
| `main.nra` | flags, planning, sharding, running, summary, exit code |
| `registry.nra` | `Suite` struct and `listSuites()`, the suite registry |
| `context.nra` | `Context`, `TestCase`, `runNarrac`, `expectRun`, `compareText`, `describeRun`, `readExpected` |
| `shell.nra` | `Command`, `runCommand`, `runScript`, `makeTempDirectory`, `removeTree`, `leakCheckEnvironment` |
| `files.nra` | sorted `listNames`, `listPaths(dir, suffix)`, `relativePath`, `replaceSuffix` |
| `diff.nra` | `describeDifference(want, got)`: first differing line, 2 lines of context, up to 8 `- want` and `+ got` lines |
| `programs.nra`, `panics.nra`, `errors.nra`, `packages.nra`, `runnertests.nra`, `bootstrap.nra` | the suites |

Suites and case counts (one command per case, all through narrac):

| Suite | Cases | Command | Pass condition |
|---|---|---|---|
| programs | 163 | `run -sanitize` or `run -release` (`test ...` for `tests.nra`, `std_test*`, `coroutine_tests*`; `-stdin x.in` when `.in` exists); `testdata/stage2/*.nra` debug only | exit 0, stdout equals `.out` |
| panics | 9 | `run -sanitize` | exit 101, stderr contains the message (table in `panics.nra`) |
| errors | 314 | `check path` | exit 1, no stdout, stderr equals the golden |
| packages | 24 | 8 packages with `run.out`: `run -sanitize`, `run -release`, `test -sanitize` | `run.out`; tests print ` passed, 0 failed` |
| runner | 2 | `test -sanitize` / `test -release` on `runner/mixed.nra` | exit 1, stdout equals `mixed.out` |
| bootstrap | 1 | `build -sanitize -o X selfhost`, then `X emit selfhost` (leak check on) and `narrac emit selfhost`, `cmp` | the two C files are equal |

Every child gets `NARRA_CHECK_LEAKS=1 ASAN_OPTIONS=detect_leaks=0`.

### Error goldens

- `testdata/errors/<name>.err` and `testdata/stage2/errors/<name>.err`: the full stderr of `narrac check <file>`.
- `testdata/packages/errors/<name>/expected.err`: the same for a package folder. It sits inside the folder because a file beside it would break Go's `Glob("errors/*")` in `TestPackageErrors`.
- 312 goldens (286 + 26) came from Go's `narra check`; narrac's output is byte-equal on all of them. The 2 stage2 goldens come from narrac (Go cannot load std modules only the self-hosted compiler has).
- The `// error:` and `# error:` first lines stay as documentation and are not read.
- `--update` rewrites a golden from narrac's output.

### How to run

```sh
NARRAC=/path/to/narrac sh scripts/test.sh              # everything
NARRAC=... sh scripts/test.sh --suite errors           # runner flags go to every shard
NARRAC=... JOBS=4 sh scripts/test.sh --filter closures --verbose
```

`test.sh` builds the runner with `$NARRAC build -release`, starts `JOBS` shards (default: CPU count) with `--shard i/JOBS` and, on a full run (no flags), one extra process for `--suite bootstrap`. It prints each log and exits 1 if any failed. `CC` selects the C compiler, as for narrac. Direct use: `testrunner [flags]`, from any directory.

### Runner flags

| Flag | Meaning |
|---|---|
| `--narrac PATH` | compiler under test (default `$NARRAC`, else `narrac`; a path with `/` is made absolute with `$PWD`) |
| `--repository DIR` | repo root (default `$PWD`); `selfhost/` for the bootstrap suite |
| `--testdata DIR` | samples (default `<repository>/compiler/testdata`); change it when testdata moves |
| `--suite NAME` / `--skip NAME` | include only / leave out a suite; repeatable; unknown name is exit 2 |
| `--filter TEXT` | keep cases whose `suite/name` contains TEXT (applied before sharding) |
| `--shard I/N` | keep cases whose position (suite order, then collect order) modulo N is I - 1; every shard sees the same list |
| `--list` | print the case names of this shard and stop |
| `--verbose` | print `ok <case>` for passes |
| `--update` | goldens are rewritten instead of compared (only the errors suite has goldens now) |

Output: `FAIL suite/name` then the message indented 4 spaces; the last line is `[shard I/N: ]P passed, F failed`. Exit 0 all passed, 1 a failure or no case matched, 2 a bad flag.

### Adding a suite (for goldens-tools and goldens-programs)

1. Write `selfhost/tests/src/<suite>.nra` exporting
   - `collectCases(context: Context): Array<TestCase> uses io, alloc`: deterministic order (sort with `Files.listPaths`), `suite` field equal to the registry name, `name` unique in the suite, `path` the sample, `variant` free text for your `runCase` (panics keeps the expected message there).
   - `runCase(context: Context, testCase: TestCase): Result<bool, string> uses io, alloc`: `Ok(true)` passes, `Err(text)` fails and the text is printed.
2. In `registry.nra` import it and add one `Suite { name, description, collectCases, runCase }` to `listSuites()`. The list order is the run order.
3. Helpers in `testrunner/context`: `runNarrac(context, arguments)` runs the compiler under test with the leak-check environment; `expectRun(result, code, stdout)`; `compareText(what, want, got)` gives the readable diff; `readExpected(path)`; `describeRun(result)`. `context.update` is the `--update` switch. Format with `narrac fmt selfhost/tests`.
4. `scripts/test.sh` needs no change; the new suite is sharded with the rest.

### How a child process runs

`Process.run` blocks, closes stdin, has no environment and no timeout, so `shell.runCommand(Command)` always runs `Process.run("sh", ["-c", script, "sh", directory, stdinPath, "NAME=value"..., program, arguments...])`. The script does `cd` (if `directory` is set), `PWD=$(pwd); export PWD` (needed by `narrac fmt`, which reads `narra.toml` through `$PWD`), uses `/dev/null` or `< stdinPath` as stdin, and `exec env "$@"`. Nothing is quoted by hand: all values are positional parameters. `Command` fields: `directory`, `environment` (`NAME=value` entries), `stdinPath`, `program`, `arguments`; `newCommand(program, arguments)` fills the rest with empty values. The result is `CommandResult { code, output, errorOutput }`; a child that cannot start has code 127. For redirections or pipelines use `runScript(script, arguments)` (see `bootstrap.nra`). Temp dirs: `makeTempDirectory()` (`mktemp -d`) and `removeTree(path)` (`rm -rf`). Stdout and stderr are captured whole; invalid UTF-8 is replaced.

## Timing

macOS arm64, 8 CPUs, narrac built by Go in release mode, cold cc:

- `sh scripts/test.sh`, 8 shards plus the bootstrap job: 33 s wall, 513 cases (shards 64 cases each).
- Go: `go test -count=1 -run 'TestPrograms$|TestPanics$|TestStringPanics$|TestCompileErrors$|TestPackages$|TestPackageErrors$|TestRunRunnerMixed$|TestStage2Programs$|TestStage2Errors$|TestSelfhostBootstrap$'` (the same samples; serial subtests; includes the sanitized stage-1/stage-2 build): 214 s.
- One shard-less run of `--suite panics --suite runner --suite packages` took 21 s serially.

## Deviations

- Error goldens for packages are `expected.err` inside the folder (reason above). Goldens for `stage2/errors` come from narrac, not Go.
- `compiler/testdata/**/*.err` adds 314 files; edited nothing else outside the owned paths except a note in `compiler/README.md` (Tests).
- The packages `tests` case uses `narrac test -sanitize` and checks ` passed, 0 failed` like the Go test; the programs that run test blocks use `narrac test` too (Go built them with `Tests` and ran them with no arguments; same stdout).
- Sample `.in` files go through `narrac run -stdin`, not through the runner's shell.
- The bootstrap case compares the sanitized compiler's `emit selfhost` with narrac's own; unlike `TestSelfhostBootstrap` it does not compare against Go (Go leaves).
- The sanitizer self-build uses narrac as stage 1: `narrac build -sanitize selfhost` already is the sanitized stage 2, so one build is enough.

## Stubs left

- None. `--update` covers only the errors suite; later suites with goldens should honor `context.update`.

## Gotchas

- The runner needs a narrac with the driver (`build`, `run`, `test`, `check`, `-sanitize`, `-stdin`, `-release`). Build it first; `scripts/test.sh` does not build narrac. Where narrac comes from without Go is for the seed-file node.
- Narra syntax met on the way: no `+` on strings (template strings only); a function-type field `f: (A) => B uses io, alloc` is the last of its list or followed by a newline-separated field (`narrac fmt` removes struct commas); narrowing `a is Err(_) || b is Err(_)` does not narrow, use one early return each.
- `Fs.listDir` has no order; use `Files.listNames`.
- A failing `narrac run` of a program that returns a number reports that number as its exit code (a panic sample that no longer panics shows e.g. exit 2).
- Shard sizes are equal by case count, not by time. Programs are the slow cases; they are first in the list.

## Proposed learnings

- A test harness that runs `narrac fmt` sets `PWD` to the working directory | the config lookup reads `$PWD` | selfhost/tests | no (done in `shell.nra`)
- Add a new suite as a module with `collectCases` and `runCase` plus one registry entry | keeps suites independent | selfhost/tests | no

## Verification

- `sh scripts/test.sh` with `NARRAC` set to a Go-built narrac: PASS, 33 s, 513 cases, under the sanitizers and the leak check, no `go` call.
- Break test: `arrays.out` first line changed, `block_expressions_bare_return.err` first line changed, `panics/bounds.nra` made not to panic. Each failed with a diff or message (for example `FAIL programs/arrays.nra/debug`, `stdout differs`, `first difference at line 1`, `- changed first line`, `+ 19 6`), and each passed after restoring the files.
- `narrac fmt selfhost/tests` applied.
- `cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`: ok, 63 s.

## Size

- Estimated M, actual M.
