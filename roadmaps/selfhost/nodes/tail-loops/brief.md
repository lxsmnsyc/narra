---
id: tail-loops
title: "Self tail calls compile to loops"
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/tail_loops*
  - roadmaps/selfhost/nodes/tail-loops/**
shared:
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/emit_call.go
  - compiler/internal/narra/emit.go
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emit.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Self tail calls compile to loops

## Goal

SPEC lines 73 and 926. `count(10000000, 0)` with a self tail call segfaults in a debug build (exit 139); the C is a plain recursive call, and release passes only because clang does tail-call optimization. A self call in tail position becomes parameter reassignment and a `goto` to the function start. Handle reference counts of the old parameter values, and mutable parameters.

## Acceptance

- [ ] A sample with 10M self tail calls runs in the debug build under ASan, with heap parameters (strings, arrays) to show the counts are right and the leak check is clean.
- [ ] Calls that are not in tail position (`return 1 + f(n - 1)`, a call inside `match` arms that then does more work) stay plain calls.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on the new samples (AST dump, types dump, emitted C) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
