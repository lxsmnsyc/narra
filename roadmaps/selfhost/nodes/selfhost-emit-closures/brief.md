---
id: selfhost-emit-closures
title: Self-hosted emitter: function values, arrow functions and environments
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
  - selfhost/src/emitclosures.nra
  - compiler/testdata/emit/skips/selfhost-emit-closures.txt
  - roadmaps/selfhost/nodes/selfhost-emit-closures/**
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

# Self-hosted emitter: function values, arrow functions and environments

## Goal

Emit function type structs and the environment runtime, thunks for named functions used as values, arrow functions with heap or stack environments and drop functions, the `// Closures` section, calls through function values, and `callFunctionValue` for the std callbacks.

## Context

- Go to port: `emit_closure.go` (230 lines).
- Captures and `stack` come from `BodyInfo.closures`. Arrow parameters are locals of the enclosing body. A generic function used as a value finds its instance through `instanceRefs`.
- `closureFunc` saves and restores the per-function state; do the same with every Emitter field the contract lists. Arrows are counted per owner from 1.

## Acceptance

- [ ] The Types section's function type part matches Go on every input.
- [ ] The skip file is deleted. The skip list shrinks by exactly the inputs that listed only this node.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 320 Narra lines in one file. The save and restore around nested bodies is the main trap.
