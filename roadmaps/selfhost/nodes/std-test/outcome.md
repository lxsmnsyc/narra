# Outcome: std-test

## Built

All in `compiler/internal/narra/` unless noted.

API (`import * as Test from "std/test"`, or named imports `import { eq } from "std/test"`):

| Function | Check | Fails with (panic message) |
| --- | --- | --- |
| `Test.eq(actual: T, expected: T)` | one type `T` (from `actual`), `T` implements `Equatable` and `Show` | `expected <Show expected>, got <Show actual>` |
| `Test.ne(actual: T, other: T)` | same | `expected a value other than <Show other>` |
| `Test.panics(f: () => void)` | `f` is `() => void` | `expected a panic, but the function returned` |
| `Test.panicsWith(f: () => void, message: string)` | same, plus a string | as above, or `expected a panic containing "<message>", got "<actual message>"` |

- A failure is an ordinary panic: stderr gets `panic at <file>:<line>:<col>: <message>`, exit code 101, and the run stops (no later tests, no `N tests passed`). `<line>:<col>` is the call's `(` (the existing `CallExpr.P` convention).
- Timeout: the child calls `alarm(N)` before `f`, N = 10 seconds or the env var `NARRA_TEST_TIMEOUT` (positive integer). Past it, the call fails with `expected a panic, but the function ran past the N s timeout`. This also bounds a nested child that keeps the pipe open. When a nested `Test.panics` times out, the inner and outer timeouts race.
- An empty `message` matches any panic, and the match stops at a NUL byte (`strstr`).
- A crash of the child (signal, sanitizer report, any exit code other than 0 or 101) fails with `expected a panic, but the function was killed by signal N` or `... exited with code N`.
- Effects: `eq`/`ne` declare `alloc`, `panics`/`panicsWith` declare `io`.
- Files: `std_testing.go` (module registration, Check and Emit), `runtime_test.h` (embedded as `Module.Runtime`), `std_test.go` (Go tests of failing assertions and of a normal build of a file that uses std/test).

Runtime hook (`runtime.h`):

- `static int narra_panic_fd = -1;` declared before `narra_panic`. When it is `>= 0`, `narra_panic` flushes stdout, writes only the message (not the position) to that fd and calls `_exit(101)`. Otherwise it behaves as before. Every panic path (`narra_panic_str`, `narra_assert`, bounds, overflow, division) goes through `narra_panic`, so all are captured.

Panic capture design (D65): fork.

- Emitted code for `Test.panics(f)`: `narra_test_proc tN; if (narra_test_fork(&tN, where)) { f.fn(f.env); narra_test_returned(); } narra_test_expect(&tN, want_or_NULL, where);`.
- `narra_test_fork` flushes stdout and stderr, makes a pipe, forks; the child closes the read end, sets `narra_panic_fd` and returns true. `narra_test_returned` flushes stdout and `_exit(0)`. `narra_test_expect` reads the pipe to EOF, `waitpid`s, and panics unless the status is exit 101 and (when given) the message contains `want` (`strstr`).
- Evidence: under clang `-fsanitize=address,undefined` with `NARRA_CHECK_LEAKS=1` and in release builds (macOS), `testdata/std_test.nra` (11 forks, one nested inside a child) and the forking cases of `std_test.go` are clean. The child's heap is a copy, so its allocations never reach the parent's leak count; `_exit` skips the child's report. setjmp/longjmp was not needed (it would skip releases and break the leak check).
- Nesting works: `Test.panicsWith(() => Test.panics(() => {}), "...")` forks inside the child.
- Output printed by the child appears in order, because the parent flushes before forking and waits.

Test-only rule:

- Every call of a `std/test` function must be inside a `test` block, arrow functions inside it included. Otherwise: E0907 `%s can only be called inside a \`test\` block` (format in `diagnostic_codes.go`). Tracked by `Checker.inTest` (`check.go`), set only while `checkBodies` checks a `TestDecl` body. Helper functions declared at top level cannot call `Test.*`, even when only tests call them.
- The import itself is allowed in any module, so a file with tests and a `run` entry still builds as a program (`TestRunStdTestBuilds`). `emit.go` now leaves out the test bodies and the std/test runtime C in a non-test build, so such a build has no fork code.

Tests:

- `testdata/std_test.nra` / `.out`: passing `eq` (int, string, char, bool, f64, struct, enum, array, Option), `ne`, `panics` (division, `panic`, bounds), `panicsWith` (message substring, failing `Test.eq`/`Test.ne`/`Test.panics` inside a closure), print in a child. `run_test.go` now runs `tests.nra` and every `std_test*.nra` in test mode.
- `std_test.go`: `TestRunStdTestFailures` (failing eq on int, struct, array; failing ne; panics whose closure returns; panicsWith with another message; an endless loop and a nested endless loop with `NARRA_TEST_TIMEOUT=1`; debug and release, exit 101, message and position, later tests not run) and `TestRunStdTestBuilds`.
- `testdata/errors/`: `std_test_outside.nra` (E0907), `std_test_no_show.nra` (`Point does not implement Show`, E0524), `std_test_no_equatable.nra` (`Map<string, int> does not implement Equatable`), `std_test_panics_type.nra` (argument count).
- Docs: SPEC.md section 13 (`std/test` bullet and example), DECISIONS.md D65, `compiler/README.md` (Supported, Tests, Not supported yet).

## Deviations

- The module is in `std_testing.go`, not the owned `std_test.go`: Go compiles a `*_test.go` file only under `go test`, so the registration would be missing from the `narra` binary. `std_test.go` holds the Go tests instead. `std_testing.go` is a new file outside the owned paths.
- The brief says importing `std/test` outside a test block or in a non-test build is an error. The checker does not know the build mode, and a file with tests must still build as a program, so the rule is on calls instead: every `Test.*` call must be inside a `test` block. Proposed in the SPEC text for the user's approval.
- `emit.go` (shared) no longer emits test bodies in a non-test build, and emits the std/test runtime only in test builds. This makes "removed from non-test builds" (SPEC section 13) true in the C. Generated C of non-test builds of files with tests changes.
- `runtime.h` gets the hook (as the brief says), although the folder rule says module C goes in `Module.Runtime`; the hook must sit inside `narra_panic`.

## Stubs left

- No Windows support for `Test.panics` (fork, alarm). Listed in `compiler/README.md` under Not supported yet.
- A failed assertion stops the whole run. Continuing after a failure, filters, reports and golden files are for test-runner.

## Gotchas

- test-runner: to run each test in isolation and keep going after a failure, the same fork pattern works per test (`narra_test_fork` / `narra_test_expect` are in `runtime_test.h`, but only emitted when a module imports std/test; a runner needs its own copy or to move them). `narra_panic_fd` makes a panic write only its message to the fd and exit 101.
- A multi-line block arrow inside call parentheses does not parse (newlines inside parentheses are not statement ends). Use a one-line block or a helper function.
- Panic positions point at the call's `(`, not at `Test`.
- `Test.eq` takes its type from the first argument; the second is checked against it. `Test.eq(None, Some(1))` fails to infer; write the typed value first.
- Failure paths free what they built except the final message passed to `narra_panic`; the leak report runs only at the normal end of `main`, so a panicking run is not leak-checked.
- Generated C for every program changed (the hook is in `runtime.h`), so C-output comparisons must regenerate.

## Proposed learnings

- Do not name a non-test Go file `*_test.go`; a std module for `std/test` lives in `std_testing.go` | Go builds such files only under `go test` | compiler/internal/narra | no
- Catch panics in test builds with a fork and `narra_panic_fd`, not setjmp | setjmp skips releases and breaks the leak check | compiler/internal/narra | yes, `TestRunStdTestFailures` and `testdata/std_test.nra` in debug mode

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (macOS, clang).
- `go test ./internal/narra -run 'TestRun|TestErrors'` (brief's verify): pass.
- By hand: `narra test testdata/std_test.nra` prints the `.out`.

## Size

- Estimated L, actual M. Fork worked on the first try under ASan and the leak check, so no setjmp spike was needed.
