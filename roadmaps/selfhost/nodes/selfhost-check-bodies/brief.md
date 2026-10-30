---
id: selfhost-check-bodies
title: Self-hosted checker: bodies
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
added: 2026-10-30
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-check-decls
    kind: hard
    why: bodies are checked against the type model and declaration tables
  - on: selfhost-check-traits
    kind: hard
    why: calls on trait functions and bounds use the trait tables
  - on: selfhost-check-imports
    kind: hard
    why: names resolve through module scopes, and builtin calls use the registry
---

# Self-hosted checker: bodies

## Goal

Port the body half of the checker to Narra: expressions, statements, patterns, closures and effects (about 4,500 Narra lines).

## Notes

- Split from selfhost-check (XL) on 2026-10-30, as the selfhost-spike outcome proposed.
- Design (user decision 2026-10-30): the parser gives every expression, pattern and local an integer id; the checker keeps its results in side tables (`Array<Type>` by id, `Map<int, ...>`). Types are a recursive `enum Type`; structs and enums are referred to by an index into a declaration table. See the selfhost-spike outcome, gap 3.
- Outline. Break into a full brief when `plan.py refine` lists it.
