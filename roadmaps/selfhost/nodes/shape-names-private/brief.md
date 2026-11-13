---
id: shape-names-private
title: "Generated Shape names are reachable from user code"
status: todo
priority: should
type: fix
story: batch-12
early: false
outline: false
milestone: 
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - testdata/errors/shape_hidden_*
  - roadmaps/selfhost/nodes/shape-names-private/**
shared:
  - selfhost/src/deriveshape.nra
  - selfhost/src/resolve.nra
  - selfhost/src/scopes.nra
reserves: []
provides: []
depends_on: []
---
# Generated Shape names are reachable from user code

## Goal

The names `derives Shape` generates (`__Shape`, `__Value`, `__ShapeText`, `__shapeToValue_<Type>`) may be reachable from user code, as `__unwrap` was before its fix. Check each, and restrict them to generated functions.

## Context

- From distinct-traits, 2026-11-12. `selfhost/src/deriveshape.nra` (HIDDEN_SHAPE, line 30) writes them.
- Proof first: a program that names one of them compiles on the base. If none is reachable, record the probes in the outcome and close the node with no code change.

## Acceptance

- [ ] Error tests for each name that was reachable.
