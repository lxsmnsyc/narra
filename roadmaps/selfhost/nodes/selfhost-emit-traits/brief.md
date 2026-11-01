---
id: selfhost-emit-traits
title: Self-hosted emitter: eq, hash, cmp and show functions and trait calls
status: todo
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
spent:
owns:
  - selfhost/src/emittraits.nra
  - compiler/testdata/emit/skips/selfhost-emit-traits.txt
  - roadmaps/selfhost/nodes/selfhost-emit-traits/**
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-emit-core
    kind: contract
    via: emitter
  - on: selfhost-emit-bodies
    kind: hard
    why: its samples compare only once plain bodies, decls and program have landed
---

# Self-hosted emitter: eq, hash, cmp and show functions and trait calls

## Goal

Emit the standard trait functions on demand (`useEq`, `useHash`, `useCmp`, `useShow`), the `// Trait functions` section, calls of compiler-implemented trait functions, and `==` and `!=` on structs, enums and arrays.

## Context

- Go to port: `emit_traits.go` (331 lines). The rules are in `contracts/trait-model.md`.
- The Self type comes from `BodyInfo.traitSelf`. User implementations come from `Checker.implFunctions`.
- The needs list grows while definitions are written; keep Go's index loop. Map functions register needs before any body (contract), so the order of needs is Go's.

## Acceptance

- [ ] The skip file is deleted. The skip list shrinks by exactly the inputs that listed only this node (traits*.nra and others).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 450 Narra lines in one file.
