---
id: std-arena
title: std/arena
status: todo
priority: should
type: feature
outline: true
milestone: p2-std
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests (std-source gate, 2026-11-05)
  - on: std-source
    kind: hard
    why: the spike decides whether this module is Narra source or compiler builtins
---

# std/arena

## Goal

`Arena<T>` and `Handle<T>` from SPEC section 7: a handle is an index plus a generation, and `Arena.get` returns `None` for a freed or reused slot. The answer for graphs and cycles under reference counting.

## Notes

Outline. Rejected before the port (DECISIONS), never scheduled after it.
