---
id: std-arena
title: std/arena
status: todo
priority: could
type: feature
outline: true
milestone: m10-std
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
  - on: std-source
    kind: hard
    why: the spike decides whether this module is Narra source or compiler builtins
---

# std/arena

## Goal

`Arena<T>` and `Handle<T>` from SPEC section 7: a handle is an index plus a generation, and `Arena.get` returns `None` for a freed or reused slot. The answer for graphs and cycles under reference counting.

## Notes

Outline. Rejected before the port (DECISIONS), never scheduled after it.
