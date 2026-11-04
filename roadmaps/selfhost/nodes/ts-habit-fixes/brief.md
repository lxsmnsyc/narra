---
id: ts-habit-fixes
title: "Diagnostics with fixes for every TypeScript habit"
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestCompileErrors|TestHabitFixes|TestErrorCodes|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/internal/narra/diagnostic_test.go
  - compiler/testdata/errors/ts_*
  - roadmaps/selfhost/nodes/ts-habit-fixes/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/lexer.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/diagnostic_fixes.go
  - compiler/internal/narra/import_names.go
  - selfhost/src/parser.nra
  - selfhost/src/lexer.nra
  - selfhost/src/resolve.nra
  - selfhost/src/modules.nra
reserves: [diag E0923, diag E0924, diag E0925, diag E0926, diag E0927, diag E0928]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Diagnostics with fixes for every TypeScript habit

## Goal

SPEC lines 165-183 and 1039, and line 191 (F16, F17). Each row of the spec's TypeScript table should give a diagnostic with a fix. Today `class`, `interface`, `new`, `try` and `throw` give generic parse errors (E0019 `expected a declaration`, E0013 `expected the end of the statement`); `for (const k in xs)` (E0031), `number` (E0113), `'abc'` (E0002) and `s.length` (E0208) have messages but `fixes=[]`; `Arrays.push(mutable xs, 2)` without the import gives E0308 `unknown module "Arrays"` with no fix, while `Array.push` gets E0908 with an add-import fix.

## Approach

Parser: recognize `class`, `interface`, `new`, `try`, `catch` and `throw` where they appear and report a dedicated diagnostic that says what Narra uses instead (struct and functions, a struct type or trait, a struct literal, `Result` and `is Err`, `panic` or `return Err(...)`), with a fix where one edit is right. Add fixes to E0031 (`in` to `of` over `Maps.keys`), E0113 (`number` to `int` or `float`), E0002 (single to double quotes) and E0208 (`s.length` to `Str.length(s)` with the import). E0308 on a std module alias that the spec's import table names (`Arrays`, `Maps`, `Str`, ...) gets the add-import fix like E0908. Keep messages short and plain.

## Acceptance

- [ ] the diagnostics tests (`TestHabitFixes` in `diagnostic_test.go`) has a case per row of the spec's TypeScript table, each with its fix in the `--json` output.
- [ ] `narra explain` covers each new code.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on every new sample and error test (AST dump, types dump, emitted C, error messages and exit codes) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
- Error tests put the position in the `// error:` header (as `testdata/errors/contains_itself.nra` does), so the test checks it.
- Earlier spec-gap nodes (block-arrow-args, box-assign, mutable-escape, tail-loops, for-destructure) changed the same parser, checker and emitter files. Several nodes of this wave edit them too: rebase on origin/main before pushing.
