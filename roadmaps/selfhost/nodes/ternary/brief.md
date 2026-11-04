---
id: ternary
title: "Conditional expression c ? a : b"
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
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/ternary*
  - compiler/testdata/errors/ternary_*
  - roadmaps/selfhost/nodes/ternary/**
shared:
  - compiler/internal/narra/lexer.go
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_effects.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/**
  - SPEC.md
reserves: [diag E0934]
provides: []
depends_on:
  - on: block-expressions
    kind: hard
    why: both add expression forms to the same parser, checker and emitter paths in both compilers; a branch may be a block expression
---

# Conditional expression c ? a : b

## Goal

Maintainer decision, 2026-11-04 (with D75). `cond ? a : b` as in TypeScript: `cond` must be `bool` (no truthiness), both branches must have the same type, and only the chosen branch is evaluated. Today `?` is `unexpected character '?'`. Same precedence and associativity as TypeScript: lower than `||`, right-associative (`a ? b : c ? d : e` is `a ? b : (c ? d : e)`).

## Acceptance

- [ ] Sample programs: a ternary as an initializer, an argument, a `return` value and inside a template; nested ternaries; branches with heap values (only the chosen branch is evaluated and released correctly; a branch with a side effect proves it); a self tail call in a returned ternary branch (`return n < 0 ? f(n + 1) : n`) stays a loop (tail-loops); a branch that is a block expression. Clean under ASan and the leak check.
- [ ] Error tests: a non-`bool` condition (with the no-truthiness message), branches of different types, a missing `:`. Positions in the `// error:` headers.
- [ ] SPEC.md section 2 (operators or control flow) describes it, and DECISIONS.md D75 gets one line that the ternary is part of it.
- [ ] The Go compiler and `selfhost/` change together and match on the new tests; fast check passes including `TestSelfhostBootstrap`.
