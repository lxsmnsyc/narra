---
id: top-level-const
title: Top-level constants
status: active
priority: should
type: feature
outline: false
milestone: m3-core
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
spent:
owns:
  - compiler/internal/narra/check_const.go
  - compiler/internal/narra/emit_const.go
  - compiler/testdata/const*
  - compiler/testdata/errors/const*
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/modules_check.go
  - compiler/internal/narra/emit.go
  - compiler/README.md
  - compiler/internal/narra/check_call.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/cnames.go
  - compiler/internal/narra/diagnostic_codes.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/modules.go
  - compiler/testdata/packages/const/**
reserves: []
provides: []
depends_on:
  - on: modules
    kind: hard
    why: constants are module members that other modules import
---

# Top-level constants

## Goal

Implement top-level `const` from `SPEC.md` section 10: a module-level constant whose value is a literal, a constructor, a struct or array of constants, a named function or another constant. Exported constants can be imported like functions. The self-hosted compiler uses them for its keyword, operator and builtin tables (selfhost-spike outcome, gap 4).

## Context

- Today `const N = 3` at the top level fails with `expected a declaration, found "const"`.
- Out of scope: `@const` functions and `lazy` (record both in `compiler/README.md` as not supported yet). A top-level value that calls a function is a compile error that says so.
- Emit constants as C static data where the type allows it (ints, floats, bools, chars, strings as static literals) and as values built once at startup otherwise. Section 10 says modules have no initialization code, so prefer static data; if an array or struct of strings or boxed enums cannot be static, build it on first use and record the difference.
- Reference counting: a constant is never freed. Make sure reads retain correctly and the leak check does not count constants as leaks.
- Diagnostic codes: take the next free codes from the table that json-diagnostics (#15) added; read its outcome for the rule. Reserve them in this brief before starting.
- Constant cycles (`const a = b; const b = a`) are an error.

## Acceptance

- [ ] `testdata/const.nra` uses int, string, struct, array and enum constants, including one imported from another module of a package, and runs clean under the sanitizers and the leak check.
- [ ] Error tests: a constant that calls a function, a cycle, assignment to a constant.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: one new declaration kind through parser, checker and emitter; static C data for the common types.
