---
id: std-test
title: std/test assertions and panic expectations
status: active
priority: should
type: feature
outline: false
milestone: m7-tooling
target:
from:
size: L
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestRun|TestErrors'
exclusive: []
fulfilled_by:
gate: true
added: 2026-10-30
owns:
  - compiler/internal/narra/std_test.go
  - compiler/internal/narra/runtime_test.h
  - compiler/testdata/std_test*.nra
  - compiler/testdata/std_test*.out
  - compiler/testdata/errors/std_test_*.nra
  - compiler/testdata/packages/std_test/**
shared:
  - compiler/internal/narra/builtins.go
  - compiler/internal/narra/driver.go
  - compiler/internal/narra/emit.go
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/runtime.h
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_effects.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/run_test.go
  - compiler/README.md
  - SPEC.md
  - DECISIONS.md
reserves: [decision D65]
provides: []
depends_on:
  - on: json-diagnostics
    kind: hard
    why: failure messages carry file:line like diagnostics
---

# std/test assertions and panic expectations

## Goal

A `std/test` module for `test` blocks: `Test.eq(a, b)` and `Test.ne(a, b)` that print both values through `Show` with the call's file and line on failure, and `Test.panics(() => ...)` (optionally with an expected message) that checks a closure panics. Today only the prelude `assert(cond, msg)` exists.

## Context

- Requested by the user on 2026-10-30 ("add a testing library `std/test` maybe"). Priority should; it runs beside the self-hosting chain, which works in `selfhost/` and touches none of this node's files except possibly `run_test.go`.
- Today: `test "name" { ... }` blocks, `narra test` (`CompileBinary(..., tests=true)` in `driver.go` emits a main that runs every block and prints `ok - <name>` and `N tests passed`, see `testdata/tests.nra`/`.out`), and the prelude `assert` (`std_prelude.go`, `narra_assert`). Panics go through `narra_panic`/`narra_panic_str` in `runtime.h` and exit with code 101.
- Std modules register through `registerModule` in `builtins.go`; follow `std_char.go` or `std_fs.go`. Every builtin sets `Builtin.Effects` (see `compiler/internal/narra/CLAUDE.md`).
- API (`import * as Test from "std/test"`):
  - `Test.eq<T: Equatable + Show>(actual: T, expected: T)` and `Test.ne`: on failure, panic with `<file>:<line>:<col>: expected <expected>, got <actual>` (show both through `Show`).
  - `Test.panics(f: () => void)` and `Test.panicsWith(f: () => void, message: string)`: fail unless calling `f` panics (with a message containing `message`).
- `std/test` is importable only in test builds; importing it from code that is not inside a `test` block, or in a non-test build, is a compile error with a new code from the format table (`TestEveryFormatHasCode`).
- Panic capture, test builds only: run each `Test.panics` closure in a child process (`fork`, run `f`, `_exit(0)` if it returns; the parent `waitpid`s and reads the panic message from a pipe). Add a runtime hook so `narra_panic` in the child writes its message to the pipe before exiting 101. POSIX only; record Windows as a gap in `compiler/README.md`. If `fork` breaks ASan or the leak check, try `setjmp`/`longjmp` around the call instead and note the leak policy; decide with evidence and write D65.
- Spec: add the module to `SPEC.md` section 13 and record the design (fork-based capture, test-only import) as D65 in `DECISIONS.md`. This node is a gate: the user approves the spec text in review.

## Acceptance

- [ ] `testdata/std_test.nra` (run with `narra test` by the harness, or a package under `testdata/packages/std_test/` if single files cannot run tests) covers passing `eq`, `ne`, `panics` and `panicsWith`, and its `.out` matches in debug (ASan, UBSan, leak check) and release builds.
- [ ] Failing cases are checked: a failing `Test.eq` prints both values and the position and the test run exits non-zero; a `Test.panics` whose closure returns fails.
- [ ] `testdata/errors/` cases: `std/test` imported outside tests; `Test.eq` on a type without `Show` or `Equatable`.
- [ ] `SPEC.md` section 13 describes `std/test`; D65 is in `DECISIONS.md`; `compiler/README.md` lists the gaps.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

L: a new std module with generic builtins, a runtime change for panic capture in test builds, a new diagnostic, and spec text.

## Notes for the worker

- Requested by the user on 2026-10-30 ("add a testing library `std/test` maybe"); priority should; scope: equality asserts, panic expectation (this node), golden files and runner features (test-runner).
- Not in `SPEC.md`. Propose the spec text in the pull request and get the user's approval; record the decision in `DECISIONS.md`.
- Panic expectation needs a way to catch a panic in test builds only. Narra has no recover (a self-hosting gap the user did not take); SPEC's `Panicked` result for coroutines is the closest model. Likely a spike inside the node: run each `Test.panics` closure in a child process or a `setjmp` frame in test builds, and check it stays clean under ASan and the leak check.
- `Test.eq` needs `Equatable` and `Show` on its type parameter.
