---
id: selfhost-emit-match
title: Self-hosted emitter: is, match, destructuring, tuples and spreads
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
  - selfhost/src/emitmatch.nra
  - compiler/testdata/emit/skips/selfhost-emit-match.txt
  - roadmaps/selfhost/nodes/selfhost-emit-match/**
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

# Self-hosted emitter: is, match, destructuring, tuples and spreads

## Goal

Emit `is` conditions, `if (x is P)`, `match` as a statement and as a value, destructuring `const`, tuple literals and spread struct literals.

## Context

- Go to port: `emit_match.go` (205 lines) and `emit_patterns.go` (54).
- Pattern variants come from `Ref.Constructor` on pattern nodes. Bound locals come from `BodyInfo.declared`. Narrowed field reads come from `Ref.Narrow`; bodies emits those, since they sit in the FieldExpr arm.
- CLAUDE.md: `&&` chains with each tag test before any read of its payload; bound names retain when bound and release when their scope pops.

## Acceptance

- [ ] The skip file is deleted. The skip list shrinks by exactly the inputs that listed only this node (patterns_*.nra, Tuple.nra and others; list them in the outcome).
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: about 350 Narra lines in one file.
