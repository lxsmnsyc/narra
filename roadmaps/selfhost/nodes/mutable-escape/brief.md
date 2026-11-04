---
id: mutable-escape
title: "A mutable parameter cannot be returned or stored"
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
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestErrors|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - compiler/testdata/errors/mutable_escape*
  - roadmaps/selfhost/nodes/mutable-escape/**
shared:
  - compiler/internal/narra/check_stmt.go
  - compiler/internal/narra/check_expr.go
  - compiler/internal/narra/diagnostic_codes.go
  - selfhost/src/stmts.nra
  - selfhost/src/exprs.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# A mutable parameter cannot be returned or stored

## Goal

SPEC line 667. `function keep(mutable p: P): P { return p }` and storing `p` in a struct literal compile clean today; capture is already rejected. Reject returning a `mutable` parameter and storing it in a literal, field, array or map. Reading it into a copy (for example `const q = p` when that copies) follows the spec text exactly; check the spec before deciding.

## Acceptance

- [ ] Compile-error tests cover `return p`, `H { p: p }`, an array literal and a field assignment, with positions on `p`. A diagnostic code is assigned in `diagnostic_codes.go`.
- [ ] `selfhost/src` still compiles: fix any place it relied on this.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on the new error tests (same messages and exit codes) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
