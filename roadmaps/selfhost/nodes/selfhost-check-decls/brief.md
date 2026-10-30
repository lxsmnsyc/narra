---
id: selfhost-check-decls
title: Self-hosted checker: declarations
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
  - on: selfhost-front
    kind: hard
    why: checks the AST the front end produces
---

# Self-hosted checker: declarations

## Goal

Port the declaration half of the checker to Narra: types, declarations, modules, generics, traits and the builtin registry (about 4,000 Narra lines).

## Notes

- Split from selfhost-check (XL) on 2026-10-30, as the selfhost-spike outcome proposed.
- Design (user decision 2026-10-30): the parser gives every expression, pattern and local an integer id; the checker keeps its results in side tables (`Array<Type>` by id, `Map<int, ...>`). Types are a recursive `enum Type`; structs and enums are referred to by an index into a declaration table. See the selfhost-spike outcome, gap 3.
- Outline. Break into a full brief when `plan.py refine` lists it.
