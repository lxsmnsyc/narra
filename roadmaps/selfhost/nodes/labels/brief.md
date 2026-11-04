---
id: labels
title: "Labeled break and continue"
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
  - compiler/testdata/labels*
  - compiler/testdata/errors/labels_*
  - roadmaps/selfhost/nodes/labels/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/stmts.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/lexer.nra
  - compiler/internal/narra/lexer.go
reserves: [diag E0916, diag E0917, diag E0918]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Labeled break and continue

## Goal

SPEC line 120 (F3). `outer: for (const i of Ranges.upTo(3)) { ... break outer }` and `outer: while (true)` fail to parse with `expected the end of the statement, found ":"`. Labels go on `for` and `while` loops; `break label` and `continue label` leave or continue that loop from any depth of nested loops.

## Approach

Parser: a name followed by `:` at the start of a statement, before `for` or `while`. Store the label on the loop statement and on `break`/`continue`. Checker: a label is in scope only inside its loop body; an unknown label, a label reused by an enclosing loop, and a label on anything but a loop are errors. Labels must not count as local names for shadowing. Emitter: `break outer` jumps past the loop, `continue outer` jumps to the loop's continue point (after the body, before the next iteration, so `for of` advances). Release locals of the inner scopes being left, as plain `break` does today.

## Acceptance

- [ ] Sample programs in `compiler/testdata/` cover `break outer` and `continue outer` from two and three nested loops, over `for of` and `while`, with heap locals in the inner scopes (clean under ASan and the leak check).
- [ ] Error tests in `compiler/testdata/errors/`: unknown label, label reused in a nested loop, label on a non-loop statement. New diagnostic codes from the reserved ones, in `diagnostic_codes.go`.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on every new sample and error test (AST dump, types dump, emitted C, error messages and exit codes) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Error tests put the position in the `// error:` header (as `testdata/errors/contains_itself.nra` does), so the test checks it.
- Earlier spec-gap nodes (block-arrow-args, box-assign, mutable-escape, tail-loops, for-destructure) changed the same parser, checker and emitter files. Several nodes of this wave edit them too: rebase on origin/main before pushing.
