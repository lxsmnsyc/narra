---
id: std-test
title: std/test assertions and panic expectations
status: todo
priority: should
type: feature
outline: true
milestone: m7-tooling
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: json-diagnostics
    kind: hard
    why: failure messages carry file:line like diagnostics
---

# std/test assertions and panic expectations

## Goal

A `std/test` module for `test` blocks: `Test.eq(a, b)` and `Test.ne(a, b)` that print both values through `Show` with the call's file and line on failure, and `Test.panics(() => ...)` (optionally with an expected message) that checks a closure panics. Today only the prelude `assert(cond, msg)` exists.

## Notes

- Requested by the user on 2026-10-30 ("add a testing library `std/test` maybe"); priority should; scope: equality asserts, panic expectation (this node), golden files and runner features (test-runner).
- Not in `SPEC.md`. Propose the spec text in the pull request and get the user's approval; record the decision in `DECISIONS.md`.
- Panic expectation needs a way to catch a panic in test builds only. Narra has no recover (a self-hosting gap the user did not take); SPEC's `Panicked` result for coroutines is the closest model. Likely a spike inside the node: run each `Test.panics` closure in a child process or a `setjmp` frame in test builds, and check it stays clean under ASan and the leak check.
- `Test.eq` needs `Equatable` and `Show` on its type parameter.
- Touches the checker hot files listed in wave 4's lessons (`check.go`, `check_call.go`, `diagnostic_codes.go`, `compiler/README.md`); list them in `shared` when refining.
- Outline. Break into a full brief when `plan.py refine` lists it.
