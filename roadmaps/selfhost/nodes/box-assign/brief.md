---
id: box-assign
title: "Assigning to b.value on a Box"
status: done
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
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestSelfhost'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
spent: 72396
owns:
  - compiler/testdata/box_assign*
  - roadmaps/selfhost/nodes/box-assign/**
shared:
  - compiler/internal/narra/emit_stmt.go
  - compiler/internal/narra/emit_box.go
  - compiler/internal/narra/check_stmt.go
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/stmts.nra
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# Assigning to b.value on a Box

## Goal

SPEC line 344. `let b = Boxes.new(5) ⏎ b.value = 6` emits invalid C: `member reference base type 'B_Box_d2ab61dd *' ... is not a structure or union` (int, struct and string payloads). Write through the box pointer, copying the cell first when its count is above 1 (copy-on-write, as for arrays). The README marks Box as supported, so this is a bug.

## Acceptance

- [ ] Sample programs assign to `b.value` for int, struct and string payloads, including a shared box (a copy made before the write leaves the other holder unchanged). Clean under ASan and the leak check.
- [ ] Assigning to `b.value` on a `const` box is a compile error with a position.
- [ ] The Go compiler and `selfhost/` change together: the self-hosted compiler matches Go on the new samples (AST dump, types dump, emitted C) through the existing `TestSelfhost*` harnesses.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`, including `TestSelfhostBootstrap` (the fixed point still holds).

## Notes

- From the spec-gap audit (`roadmaps/selfhost/nodes/spec-gap-audit/outcome.md`): its findings table has the minimal program and the exact error.
