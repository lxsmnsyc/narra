---
id: test-runner
title: Golden files and narra test runner features
status: done
spent: 510000
priority: should
type: feature
outline: false
milestone: m7-tooling
target:
from:
size: M
spike: false
verify: cd compiler && go test ./internal/narra -run 'TestRun|TestErrors'
exclusive: []
fulfilled_by:
gate: true
added: 2026-10-30
owns:
  - compiler/internal/narra/test_runner.go
  - compiler/internal/narra/runner_test.go
  - compiler/testdata/runner/**
  - compiler/testdata/packages/runner/**
shared:
  - compiler/cmd/narra/main.go
  - compiler/internal/narra/driver.go
  - compiler/internal/narra/emit.go
  - compiler/internal/narra/runtime.h
  - compiler/internal/narra/runtime_test.h
  - compiler/internal/narra/std_testing.go
  - compiler/internal/narra/std_test.go
  - compiler/internal/narra/run_test.go
  - compiler/internal/narra/modules_test.go
  - compiler/testdata/tests.out
  - compiler/testdata/std_test.out
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_effects.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/README.md
  - SPEC.md
  - DECISIONS.md
reserves: [decision D66]
provides: []
depends_on:
  - on: std-test
    kind: hard
    why: golden comparison reports failures like Test.eq, and the runner counts std/test failures
---

# Golden files and narra test runner features

## Goal

`Test.golden(name, actual)` compares a string with a file under the package's `testdata/` and rewrites it when `NARRA_UPDATE_GOLDEN=1` is set. `narra test` gains `--filter <text>`, per-test timing and a summary of passed, failed and total tests.

## Context

- std-test (#24) is merged: read its outcome. A failed assertion is a panic that stops the whole run. `runtime_test.h` has `narra_test_fork` / `narra_test_expect` (fork, pipe, `alarm` timeout via `NARRA_TEST_TIMEOUT`), emitted only when a module imports std/test; `narra_panic_fd` in `runtime.h` makes a panic write its message to a pipe and exit 101. Test bodies are emitted only in test builds (`emit.go`).
- The test main today prints `ok - <name>` per test and `N tests passed` (`testdata/tests.out`).
- Runner: run each test in its own child with the same fork pattern (move the fork helpers so every test build has them, not only builds importing std/test), keep going after a failure, print `ok - <name>` or `FAIL - <name>: <message> (<file>:<line>:<col>)`, then a summary `N passed, M failed` and exit 1 when any failed. Per-test timing in milliseconds behind `--time` (timing must not appear in compared output by default). `narra test --filter <text>` runs only tests whose name contains the text. The runner reads the filter from argv or an env var set by `narra test`; pick one and document it.
- `Test.golden(name: string, actual: string)`: compare with `<package>/testdata/<name>.golden` (single files: next to the file); with `NARRA_UPDATE_GOLDEN=1` write the file instead. A mismatch fails with the first differing line. Decide how it may do file I/O inside a test block under effect checking (`Test.golden` declares `io`; check that test blocks accept `io`). Same test-only rule as the other `Test.*` calls (E0907).
- Windows stays unsupported (fork); keep the README gap.
- Spec: extend SPEC section 13 (runner output, `--filter`, `--time`, `Test.golden`) and record the design as D66. Gate: the user approves the spec text.

## Acceptance

- [x] A sample with passing and failing tests runs every test, prints `ok`/`FAIL` lines and the summary, exits 1, in debug (ASan, UBSan, leak check) and release builds; an all-passing run exits 0. Existing `tests.nra` and `std_test*.nra` outputs are updated to the new summary format.
- [x] `--filter` and `--time` work and have tests; timing never appears in compared output.
- [x] `Test.golden` passes on a match, fails with the first differing line on a mismatch, and rewrites the file under `NARRA_UPDATE_GOLDEN=1` (test in a temp dir).
- [x] `SPEC.md` section 13, D66 and `compiler/README.md` describe it.
- [x] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: a test main rewrite around the existing fork helpers, two CLI flags and one builtin.

## Notes for the worker

- Requested by the user on 2026-10-30, with std-test. Golden files would let the self-hosted compiler test its own dumps in Narra.
- Not in `SPEC.md`. Propose the spec text in the pull request and get the user's approval.
- Golden files need `uses io` in tests; decide whether test blocks may use `io` freely.
