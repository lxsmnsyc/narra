---
id: let-narrowing
title: "Narrowing a let after an is test"
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
  - compiler/testdata/let_narrowing*
  - compiler/testdata/errors/let_narrowing_*
  - roadmaps/selfhost/nodes/let-narrowing/**
shared:
  - compiler/internal/narra/check_patterns.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/check_expr.go
  - selfhost/src/patterns.nra
  - selfhost/src/stmts.nra
  - selfhost/src/exprs.nra
  - selfhost/src/match.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Narrowing a let after an is test

## Goal

SPEC lines 407 and 569 (F9). `let x: Option<int> = Some(1) ⏎ if (x is None) { return 0 } ⏎ print(Int.toString(x.value))` fails with `this Option<int> may hold any of its variants`. Narrowing works on `const` and parameters only. A `let` must narrow the same way until its next assignment.

## Approach

Checker only. Narrowing facts for a `let` end at any assignment to it (`x = ...`, a field or index assignment, `mutable x` passed to a call) and at the start of a loop body that assigns it anywhere (the next iteration can see a changed value). Closures cannot capture `let` today, so they cannot change it. Keep the existing `const` and parameter paths unchanged.

## Acceptance

- [ ] Sample programs cover `is None` early return then `x.value`, `is Err(e)` early return then `r.value`, narrowing inside an `if (x is Some(v))` branch, and a loop that reassigns the `let` (narrowing must not survive into the next iteration).
- [ ] Error tests: reading `.value` after an assignment that ends the narrowing, and after passing `mutable x` to a call.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on every new sample and error test (AST dump, types dump, emitted C, error messages and exit codes) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Error tests put the position in the `// error:` header (as `testdata/errors/contains_itself.nra` does), so the test checks it.
- Earlier spec-gap nodes (block-arrow-args, box-assign, mutable-escape, tail-loops, for-destructure) changed the same parser, checker and emitter files. Several nodes of this wave edit them too: rebase on origin/main before pushing.
