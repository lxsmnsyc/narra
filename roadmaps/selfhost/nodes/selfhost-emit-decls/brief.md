---
id: selfhost-emit-decls
title: Self-hosted emitter: struct, enum, array and box declarations and reference counting
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
spent: 196108
owns:
  - selfhost/src/emitdecls.nra
  - compiler/testdata/emit/skips/selfhost-emit-decls.txt
  - roadmaps/selfhost/nodes/selfhost-emit-decls/**
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
---

# Self-hosted emitter: struct, enum, array and box declarations and reference counting

## Goal

Emit the Types and Reference counting sections: struct and enum definitions, array structs and functions, retain and release functions with the `narra_drop` worklist, Box and enum payload heap types, and `enumAllocation` for the other nodes.

## Context

- Go to port: `emit_types.go` (191 lines: `typeDef`, `rcFuncs`, `arrayFuncs`) and `emit_box.go` (100 lines: `typedefHeap`, `heapDefs`, `boxProtos`, `boxFuncs`, `enumAlloc`). The hook order is in the contract.
- CLAUDE.md rules for enums with fielded variants (`{ tag; Q* u }`, build only through `enumAllocation`) and for the drop worklist apply as they do in Go.

## Acceptance

- [ ] The Types section (except the map and closure parts) and the Reference counting section (except the map prototypes) match Go on every input.
- [ ] The skip file is deleted. The skip list shrinks by exactly the inputs that listed only this node.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 400 Narra lines in one file, with string templates mirrored from Go.
