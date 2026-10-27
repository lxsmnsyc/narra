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

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
