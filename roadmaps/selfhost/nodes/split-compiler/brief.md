---
id: split-compiler
title: Split the checker and emitter by feature
status: done
priority: must
type: refactor
outline: false
milestone: m3-core
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: true
added:
owns:
  - compiler/internal/narra/check*.go
  - compiler/internal/narra/emit*.go
  - compiler/internal/narra/builtins.go
  - compiler/internal/narra/std_int.go
  - compiler/internal/narra/std_float.go
  - compiler/internal/narra/std_string.go
  - compiler/internal/narra/std_prelude.go
  - compiler/internal/narra/diagnostic.go
shared: []
reserves: []
provides: [diagnostic, builtins]
depends_on: []
---

# Split the checker and emitter by feature

## Goal

Split `check.go` and `emit.go` into one file per feature, move standard library functions into a registry with one file per module, and route every compile error through a `Diagnostic` value. No behavior changes. This lets later nodes add features in their own files instead of all editing the same two files.

## Context

`check.go` (about 1,400 lines) and `emit.go` (about 900 lines) hold every feature. Almost every node in this roadmap edits them, so they would serialize the whole plan. Builtins today are a `map` in `check.go` plus a `switch` in `emit.go`.

- Source: `SPEC.md`, `DECISIONS.md`, `compiler/README.md`

## Steps

- [ ] Split the checker into `check.go` (program, scopes, entry), `check_types.go` (type resolution and generics), `check_stmt.go`, `check_expr.go`, `check_call.go`, `check_match.go` (patterns and `is`).
- [ ] Split the emitter into `emit.go` (program layout), `emit_types.go` (type definitions, reference counting, arrays), `emit_stmt.go`, `emit_expr.go`, `emit_call.go`, `emit_match.go`.
- [ ] Add the builtin registry from `contracts/builtins.md` and move `std/int`, `std/float` and `std/string` into `std_int.go`, `std_float.go` and `std_string.go`, and `print`, `assert`, `panic`, `Array.*` and `Range.*` into `std_prelude.go`. `for ... of` keeps its special cases in the statement code.
- [ ] Add per-module runtime C: a module registers its C helpers, and the emitter includes them only when the module is imported (see `contracts/builtins.md`).
- [ ] Add the `Diagnostic` type from `contracts/diagnostic.md`. `errorf` calls `report`. Wrap lexer and parser errors in one `Diagnostic`. Keep the printed text and order the same.
- [ ] Mark both contracts `status: frozen` in the outcome.

## Acceptance

- [ ] Existing tests pass unchanged; the diff only moves code, apart from the registry and `Diagnostic` types.
- [ ] Adding a new std function needs no edit to `check_call.go` or `emit_call.go`.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: many files, but the change moves code without changing behavior, and the tests already cover it.

## Notes for the worker

- Do not rename functions or change logic in the same pull request. Reviewers should be able to check that code only moved.

Pull request: #2 (merged 2026-10-28 by the user).
