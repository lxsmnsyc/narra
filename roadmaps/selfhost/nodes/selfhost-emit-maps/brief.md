---
id: selfhost-emit-maps
title: Self-hosted emitter: Map types, literals, indexes, loops and std/map
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
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostEmit'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - selfhost/src/emitmaps.nra
  - compiler/testdata/emit/skips/selfhost-emit-maps.txt
  - roadmaps/selfhost/nodes/selfhost-emit-maps/**
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
  - on: selfhost-emit-traits
    kind: behavior
    via: emitter
---

# Self-hosted emitter: Map types, literals, indexes, loops and std/map

## Goal

Emit map typedefs, structs, prototypes and functions from the `runtime_map.h` template, Map literals, `m[k]` reads and writes, `for` over `Map.keys` and `Map.entries`, and the std/map builtins.

## Context

- Go to port: the emitter half of `std_map.go` (lines 27-53 Emit closures, 254-404; about 200 lines). The template text is `RUNTIME_MAP_H` in runtimetext.nra.
- A write is an AssignStmt whose target is an index on a `TMap`; Narra records no `MapSet` flag.

## Acceptance

- [ ] The map parts of the Types and Reference counting sections match Go on every input.
- [ ] The skip file is deleted. The skip list shrinks by exactly the inputs that listed only this node (maps.nra and others).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

S: about 250 Narra lines in one file, with one template substitution.
