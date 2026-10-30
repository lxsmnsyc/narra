# Outcome: test-runner

## Built

All in `compiler/internal/narra/` unless noted.

- `test_runner.go`: `emitTestMain` writes the C main of a test build (case table plus `narra_test_run`), `TestArgs(filter, timing)` turns the `narra test` flags into test-binary arguments, `goldenDir(pkg)` gives the golden directory.
- Runner behavior: each test runs in a forked child. The child sets `narra_panic_fd` and the new `narra_panic_pos`, runs the test, frees constants, runs the leak check and exits. The parent prints `ok - <name>` or `FAIL - <name>: <message> (<file>:<line>:<col>)` (signal and exit-code failures have their own text), then `N passed, M failed`; exit 1 on any failure. Stdout is flushed before each fork, so test output stays in order.
- Options are argv of the test binary: `--filter <text>` / `--filter=<text>` (substring of the name) and `--time` (appends ` (N ms)` to each line). `narra test -filter text -time file` passes them on (`cmd/narra/main.go`; both flags are rejected for build and run).
- `Test.golden(name, actual)` in `std_testing.go` (effect `io`, test-only, E0907 outside a test) with its C in `runtime_test.h`. It compares with `<dir>/<name>.golden`, writes it under `NARRA_UPDATE_GOLDEN=1` (creating directories), and fails with `golden <name> differs at line N: expected "<line>", got "<line>"` or a missing-file message. Dir: `<package>/testdata`, or the file's directory for a single file; absolute, baked in at compile time through the new `Program.GoldenDir` (`check.go`, set in `driver.go`).
- Test blocks accept `io` (and any effect) already: the effect pass discards the effects of a `test` body. No change needed; documented in D66.
- `runtime.h`: `narra_panic_pos`; with it set, `narra_panic` writes `<where>\n<message>` to the pipe. `narra_test_fork` resets it in nested children.
- Tests: `runner_test.go` (`TestRunRunnerMixed`, `...FilterAndTime`, `...Golden` in a temp dir, `...GoldenPackage`, `...GoldenOutsideTest`), samples `testdata/runner/mixed.nra` + `.out`, `testdata/packages/runner/` (golden in a package). `tests.out` and `std_test.out` use the new summary.
- Docs: SPEC.md section 13, DECISIONS.md D66, `compiler/README.md`.

## Deviations

- `std_test.go` (shared): `TestRunStdTestFailures` now expects exit 1 and the `FAIL` line, and that the next test runs.
- `modules_test.go` (not owned, not shared): one-line change, the package test check looks for ` passed, 0 failed` instead of `tests passed`.
- Failed assertions: exit code is 1 (was 101), message is on stdout in the `FAIL` line (was `panic at ...` on stderr). Panics outside tests are unchanged.
- `Test.golden` goes into `std_testing.go` / `runtime_test.h` (std module rule), not into `test_runner.go`.
- The runner's C is a Go string in `test_runner.go`, emitted after the program's functions (it calls `narra_consts_release`). `runtime_test.h` is still emitted only when std/test is imported; the runner does not need it.

## Stubs left

- Windows: fork, so unsupported (README gap kept).
- No JSON or other report formats.

## Gotchas

- The leak check now runs per test child; a leak fails that test with `leaked N allocations.
- A test using `Test.golden` in update mode passes and rewrites files; run it only on purpose.
- `narra test` passes `--golden-dir`; a test binary run alone reads `testdata/` relative to its working directory. No path is in the generated C.
- The ordering `FAIL` line vs. test output: output printed by the test comes first, as lines before the FAIL line.
- Generated C of every test build changed.

## Proposed learnings

- Run each test block in its own child via `narra_test_run`; do not put test logic in the generated main | isolation and leak check per test | compiler/internal/narra | yes, `TestRunRunnerMixed`
- Do not name the `narra test` summary differently without updating `modules_test.go` and the `.out` files | they compare it | compiler/testdata | yes

## Verification

- `cd compiler && go vet ./... && go test ./...`: pass (macOS, clang; debug under ASan/UBSan and leak check, and release).
- Brief's verify, `go test ./internal/narra -run 'TestRun|TestErrors'`: covered by the full run, pass.
- By hand: `narra test testdata/runner/mixed.nra` (exit 1), `-filter passes -time`.

## Size

- Estimated M, actual M.
