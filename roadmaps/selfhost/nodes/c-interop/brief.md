---
id: c-interop
title: Unsafe code and C interop
status: todo
priority: should
type: feature
outline: true
milestone: m4-systems
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
  - on: modules
    kind: behavior
    via: module-graph
---

# Unsafe code and C interop

## Goal

`unsafe` blocks, `Ptr<T>`, `extern`, `@repr(C)`, `@repr(u8)` enums, `@export(C)`, `CString`, `[link]` in `narra.toml`, and `narra audit`.

From hashed-cnames: `__` and edge `_` are now allowed in names, so `@export(C)` and `extern` names are the only C names that are not hashed. They must be checked against the name table so they cannot clash with a generated name. From box-recursive: a fielded enum is `{ tag, Q* u }`, so `@repr(C)` on such an enum needs its own rule, and `@repr(u8)` fits only enums without fields.


## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
