---
id: selfhost-emit-bodies
title: Self-hosted emitter: functions, statements, expressions, calls and constants
status: done
priority: must
type: feature
story:
early: false
outline: false
milestone: m8-selfhost
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent: 440730
owns:
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitconsts.nra
  - compiler/testdata/emit/skips/selfhost-emit-bodies.txt
  - roadmaps/selfhost/nodes/selfhost-emit-bodies/**
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-emit-core
    kind: contract
    via: emitter
  - on: selfhost-emit-program
    kind: behavior
    via: emitter
  - on: selfhost-emit-decls
    kind: behavior
    via: emitter
---

# Self-hosted emitter: functions, statements, expressions, calls and constants

## Goal

Emit function and test bodies: every statement, every plain expression, calls of user functions (generic instances through `instanceRefs`, impl functions), enum constructors, distinct calls, the prelude and Array, Range and Box builtins, `for` over ranges and arrays, and top-level constants. All dispatch arms to the feature modules go in as the contract lists them.

## Context

- Go to port: `emit_expr.go` (242 lines), `emit_stmt.go` (167), `emit_call.go` (60), `emit_const.go` (104), `function`/`blockInScope`/`block` and the test bodies from `emit.go` (296-323, 116-131), and the Emit halves of `std_prelude.go`, `std_array.go`, `std_range.go` and `std_box.go`.
- Static constants re-emit their value in the current function, so read their types through `switchBody`. Field defaults do the same with the default body.

## Rules

- Each dispatch arm to a feature module is one line, with the contract's signature. Do not implement feature constructs here.
- Keep Go's order of temps, literals and releases. Every counter shows in the output.

## Acceptance

- [ ] The Constants section matches Go on every input. Every input that listed only this node (with decls merged) compares fully.
- [ ] The skip file is deleted. The skip list shrinks by exactly those inputs. The outcome lists them.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M (upper end): about 850 Narra lines. The code is a direct port, but it is the widest emitter file set.

## Notes for the worker

- Integer literals: Go prints `INT64_C(%d)` of `ParseInt(text, 0, 64)` and `INT64_MIN` for the minimum. Floats print their source text; chars print as `%dU`.
- An ExprStmt prints only void calls. A call that returns a value goes through `temp` (CLAUDE.md).
- Recursion on expressions is fine. Loop over statements.
