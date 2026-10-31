---
id: selfhost-emit
title: Self-hosted C emitter
status: todo
priority: must
type: feature
outline: true
milestone: m8-selfhost
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-check-bodies
    kind: hard
    why: emits from the checked tree; per-body type tables and the dump format
  - on: selfhost-check-calls
    kind: hard
    why: emits from the checked tree; call resolution and generic instances
  - on: selfhost-check-patterns
    kind: hard
    why: emits from the checked tree; patterns and narrowing
  - on: selfhost-check-closures
    kind: hard
    why: emits from the checked tree; closure captures and stack environments
  - on: selfhost-check-builtins
    kind: hard
    why: emits from the checked tree; builtin calls
  - on: selfhost-check-effects
    kind: hard
    why: emits from the checked tree; the complete checker
  - on: selfhost-names-exported
    kind: hard
    why: reads the AST and checker fields under their final names
---

# Self-hosted C emitter

## Goal

C emitter and driver in Narra, producing the same C as the Go compiler.

## Notes

- 2026-10-30: port the xxHash32 check that two modules' relative paths do not hash alike (`module-graph` contract); selfhost-modules deferred it here.

Outline. Break into full briefs when `plan.py refine` lists it.
