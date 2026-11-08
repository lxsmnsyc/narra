# Outcome: selfhost-driver

## Built

`narrac build`, `run`, `test` and `check` in `selfhost/src/driver.nra` (new module `narrac/driver`). `main.nra` dispatches with `Driver.isDriverCommand(arguments[1])` then `Driver.runDriver(arguments)`, before every older command. Nothing of this existed before. `arguments[0]` is the program name.

Flags of build, run and test follow Go's `flag` syntax: `-name`, `--name`, `-name=value`, `-name value`; the first non-flag argument ends the flags (for run and test the rest are program arguments); `--` ends them and is dropped; `-h` and `-help` print the usage and exit 2 (Go exits 0).

| flag | meaning |
|---|---|
| `-release` | `-O2 -DNARRA_RELEASE` (else `-O0 -g`) |
| `-o path` | the executable. build of a package: one binary only. run and test: keep the executable here instead of a temp one |
| `-emit-c file` | also keep the generated C |
| `-bin name` | package binary to build or run |
| `-filter text`, `-time` | test only; with build or run: message, exit 2 |
| `-sanitize` | adds `-fsanitize=address,undefined -fno-omit-frame-pointer` (extension, Go has none) |
| `-stdin file` | run and test: the program reads the file as stdin, through `sh -c 'input=$1; shift; exec "$@" < "$input"'` (extension) |

Commands:
- `narrac build [flags] file.nra|dir`: a file builds `-o` or the basename without `.nra`; a package builds every `[[binary]]` (named after the binary), or `-bin`. Errors: `-o needs one binary; ...`, `narra.toml: package X has no [[binary]]; it is a library`.
- `narrac run [flags] file.nra|dir [args...]`: builds in a temp dir, runs, exits with the program's exit code (101 for a panic).
- `narrac test [flags] file.nra|dir [args...]`: builds the test program (no entry function needed) and runs it with `[--filter X] [--time] --golden-dir <abs>` then the extra args. Golden dir is the directory of the file, or `<dir>/testdata` for a package, made absolute with `$PWD`.
- `narrac check file.nra|dir`: load, body check, effects check (effects only when bodies have no errors), then a diagnostic for each binary without an entry function. Human form on stderr: `file:line:col: message` per line. There is no `--json` (see Deviations); any flag gives the usage and exit 2.

Exit codes: 0 success; 1 load, check, entry, unsupported-construct or C compiler error (build), or the program's own exit code (run, test); 2 usage (no path, unknown or malformed flag, wrong flag for the command, `check` without exactly one path or with a flag).

cc invocation (`$CC` if set and non-empty, else `cc`), via `Process.run`:
- debug: `cc -O0 -g -std=c11 -w [-fsanitize=address,undefined -fno-omit-frame-pointer] -o OUT out.c -lm`
- release: `cc -O2 -DNARRA_RELEASE -std=c11 -w [sanitize flags] -o OUT out.c -lm`
- cc stdout and stderr both go to stderr. Failure: `C compiler failed: exit status N`, exit 1. The leak check is the caller's environment: set `NARRA_CHECK_LEAKS=1` (and `ASAN_OPTIONS=detect_leaks=0`) before running `narrac run` or `test`; the child inherits them.

Also: `findEntry` in `emitter.nra` is now `export`ed; the driver reuses it.

## Proof

Compared `narra-go` and `narrac` (built by Go, `-release`) on stdout, stderr and exit code (script in the session scratchpad): 9 programs with `run`, 12 error samples with `check` and `build`, `runner/mixed.nra` with `test` and `test -filter`, `testdata/tests.nra`, `selfhost` with `check` and `test`, all package samples with `check`, `run` and `build`, panics (exit 101), `-emit-c`, `-o`, `-release` with program args, missing path and nonexistent file. All identical, except:
- `check --json` was left out.
- usage and flag error text differs (our own usage; exit code 2 equal).
- a missing `narra.toml` message text (runtime error text).
Extensions `-stdin` and `-sanitize` were run by hand (stdin sample prints the right bytes).

## Deviations

- `Process.run` captures output and closes stdin: `run`/`test` print the program's stdout, then its stderr, after it ends (no interleaving, no streaming, no interactive stdin, invalid UTF-8 is replaced). `-stdin` covers the test inputs.
- Added `-sanitize` and `-stdin`. Go has neither (its sanitizer flags live in the test harness).
- `check --json` is left out, as the coordinator directed: the self-hosted checker has no diagnostic codes or JSON diagnostics (Go: `diagnostic_codes.go`). The next node adds them.
- Coordinator correction applied: debug `build`, `run` and `test` use no sanitizers by default (as Go); `-sanitize` adds exactly `-fsanitize=address,undefined -fno-omit-frame-pointer`. The brief's "debug with ASan and UBSan" was wrong. Single-dash and double-dash flags are both accepted (`-release`, `--release`, `--sanitize`). The leak check stays `NARRA_CHECK_LEAKS=1` in the environment.
- Edited outside the owned paths: `selfhost/src/emitter.nra` (`export` on `findEntry`, one word).
- Error text of the message-only cases says `narrac`, not `narra`.

## Stubs left

- `check --json` and diagnostic codes (next node).
- Passthrough of stdio for child processes (needs a runtime change) – change request below.

## Gotchas

- `Process.run` finds a name without a slash on PATH; the driver prefixes `./`.
- `mktemp -d` makes the temp dirs; removal uses `removeFile` and `removeDir` (rmdir), so only files the driver put there.
- Syntax errors from `Modules.load` arrive as one message starting with `error `, without file or position; the driver strips the prefix.
- Single-file `check` errors print the file basename as Go does, because the checker stores it that way.

## Proposed learnings

- Run children that need stdin or live output through a runtime passthrough, not Process.run | no passthrough exists | selfhost/ | no

## Verification

- `cd compiler && go test -timeout 45m ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`: 110 passed.
- The new files pass `narrac fmt` (applied before the tests).

## Size

- Estimated M, actual M.
