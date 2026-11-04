---
id: unused-names
title: "Unused locals and parameters are errors"
status: todo
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestHabitFixes|TestErrorCodes|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/errors/unused_*
  - roadmaps/selfhost/nodes/unused-names/**
shared:
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/check_patterns.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/diagnostic_fixes.go
  - compiler/testdata/**
  - selfhost/src/**
  - bench/**
  - compiler/README.md
reserves: [diag E0921, diag E0922]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
  - on: let-narrowing
    kind: hard
    why: both rewrite how the checker tracks locals in check_stmt.go; unused-names also cleans every sample, so it goes after the other checker nodes
  - on: labels
    kind: hard
    why: unused-names cleans every sample and selfhost file; run it after the nodes that add samples so they are cleaned once
---

# Unused locals and parameters are errors

## Goal

SPEC line 57 (F5). Unused locals and parameters compile clean. The maintainer chose the spec over lint (2026-11-04): an unused local or parameter is a checker error with a fix. Read SPEC.md line 57 and nearby for the exact rule (how to mark a name as intentionally unused, for example a `_` prefix or `_` alone, and whether pattern bindings, loop variables and parameters of implement-block functions count).

## Approach

Checker: track reads per local; at the end of its scope, a local never read is an error whose fix renames it to the spec's unused form or removes it. Parameters the same, except where the spec exempts them (for example functions that must match a trait signature). The check must not fire on hidden names made by desugarings (for-destructure's `__for_<line>_<column>`). First run the new check over `compiler/testdata/`, `selfhost/src/`, `bench/` and the samples in SPEC.md, and clean every hit in a separate first commit, so the rule commit is small. If cleaning `selfhost/src` alone grows past M, stop and report the count.

## Acceptance

- [ ] Every existing sample, `selfhost/src` and `bench/` compile under the new rule; the cleanup is its own commit.
- [ ] Error tests cover an unused `const`, `let`, parameter, pattern binding and loop variable, each with the fix in `--json` output (`TestHabitFixes` in `diagnostic_test.go`).
- [ ] Intentionally unused names in the spec's form compile clean; hidden desugared names never report.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on every new sample and error test (AST dump, types dump, emitted C, error messages and exit codes) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Error tests put the position in the `// error:` header (as `testdata/errors/contains_itself.nra` does), so the test checks it.
- Earlier spec-gap nodes (block-arrow-args, box-assign, mutable-escape, tail-loops, for-destructure) changed the same parser, checker and emitter files. Several nodes of this wave edit them too: rebase on origin/main before pushing.
