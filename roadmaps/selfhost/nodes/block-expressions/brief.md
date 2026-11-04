---
id: block-expressions
title: "Block expressions and if/else as a value"
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/block_expressions*
  - compiler/testdata/errors/block_expressions_*
  - roadmaps/selfhost/nodes/block-expressions/**
shared:
  - compiler/internal/narra/parser.go
  - compiler/internal/narra/ast.go
  - compiler/internal/narra/clone.go
  - compiler/internal/narra/astdump.go
  - compiler/internal/narra/check.go
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/check_closure.go
  - compiler/internal/narra/check_effects.go
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/emit_expr.go
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/**
  - SPEC.md
  - compiler/README.md
reserves: [diag E0931, diag E0932, diag E0933]
provides: []
depends_on:
  - on: let-narrowing
    kind: hard
    why: both change statement checking in check_stmt.go and stmts.nra; build on the merged narrowing rules
  - on: for-of-sources
    kind: hard
    why: both change the for statement in both compilers; build on the merged version
---

# Block expressions and if/else as a value

## Goal

DECISIONS.md D75 (maintainer, 2026-11-04). A block in value position is a block expression: `const label = { const name = lookup(code) ⏎ if (name == "") { return "unknown" } ⏎ return name }`. Its value comes from `return v`. A `return` anywhere inside it (also in nested block statements, loops and `match` arms) ends the block expression, not the function; `break` and `continue` cannot leave it. `if (c) { ... } else { ... }` in value position is an expression whose branches are block expressions. A block in statement position stays a block statement, so every existing program keeps its meaning.

## Approach

Value positions: the initializer of `const`/`let`, the right side of an assignment, a call argument, an operand, a `return` value, a template `${...}`, an array, tuple or struct-field element. Parse `{` there as a block expression, and `if` there as an if-expression (an `else` branch is required; `else if` chains are allowed). In the checker, a block expression is a new return target: the type of its value is the common type of its `return`s; every path must end in `return` (E0931, like a function missing a return); `return` with no value inside one is an error; `break`/`continue` to a loop outside it is E0932; a block expression with no `return` at all, such as TypeScript's `const o = { a }`, gets E0933 with a message that Narra has no object literals and points to struct literals. Narrowing, effects, mutable rules and closure capture see the block as part of the enclosing function. Emitter: a C block with a result temporary and a `goto` to its end for each `return`; release the locals of the scopes being left, as `break` does. A self tail call inside a block expression is not a tail call of the function.

## Acceptance

- [ ] Sample programs cover a block expression as an initializer, an argument and an operand; a nested block statement and a loop with an early `return` inside it; heap locals released on every path; `if`/`else if`/`else` as a value with block and single-expression branches; a block expression inside an arrow body. Clean under ASan and the leak check.
- [ ] Error tests (positions in the `// error:` headers): a path without `return` (E0931), `break outer` across the block (E0932), `const o = { a }` (E0933), `return` with no value, branches of different types, `if` as a value without `else`.
- [ ] Existing block statements, `if` statements and `match` arms with blocks behave as before (all current samples pass unchanged).
- [ ] SPEC.md section 2 describes block expressions and if-expressions with the rules above; "Differences from TypeScript" gets a row for `const o = { a }`.
- [ ] The Go compiler and `selfhost/` change together; the self-hosted compiler matches Go on every new sample and error test through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap`.

## Notes

- match-block-values (next node) makes `match` arms with blocks produce values, on top of this node.
- Use your own scratch directory under the session scratchpad; other agents share it.
