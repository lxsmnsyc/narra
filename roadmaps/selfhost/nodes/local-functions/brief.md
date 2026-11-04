---
id: local-functions
title: "Functions declared inside functions"
status: active
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/local_functions*
  - compiler/testdata/errors/local_functions_*
  - roadmaps/selfhost/nodes/local-functions/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/check_closure.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/stmts.nra
  - selfhost/src/declare.nra
  - selfhost/src/closures.nra
reserves: [diag E0919]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Functions declared inside functions

## Goal

SPEC line 62 (F2). `function twice(x: int): int { return x * 2 }` inside a function body fails with `expected an expression, found "function"`. Read SPEC.md around line 62 for the exact rules (whether a local function may capture, recursion, use before declaration) and follow them; record any gap in the outcome.

## Approach

Preferred: a local function that captures nothing desugars to a private module-level function with a hidden name built from the enclosing function and position (as for-destructure builds `__for_<line>_<column>`), so checker and emitter need no new case. Both compilers must produce the same hidden name and the same C. If the spec lets a local function capture locals, a capturing one is a closure bound to a `const` (the existing arrow-function path), including the recursion rules. If capture needs more than M, implement non-capturing local functions, make capture a clear compile error, and report the rest as a follow-up node.

## Acceptance

- [ ] Sample programs cover a local function called before and after its declaration (as the spec says), a recursive local function, two local functions with the same name in different functions, and (if the spec allows it) one that captures a `const`.
- [ ] Error tests cover the cases the spec rejects (for example a name that shadows a local or a module function).
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on every new sample and error test (AST dump, types dump, emitted C, error messages and exit codes) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Error tests put the position in the `// error:` header (as `testdata/errors/contains_itself.nra` does), so the test checks it.
- Earlier spec-gap nodes (block-arrow-args, box-assign, mutable-escape, tail-loops, for-destructure) changed the same parser, checker and emitter files. Several nodes of this wave edit them too: rebase on origin/main before pushing.
