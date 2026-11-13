---
id: std-unicode
title: std/unicode
status: todo
priority: should
type: feature
story: batch-14
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
  - on: std-design
    kind: hard
    why: the design gate fixes the API this node builds
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests (std-source gate, 2026-11-05)
  - on: std-source
    kind: hard
    why: the spike decides whether this module is Narra source or compiler builtins
---

# std/unicode

## Goal

Grapheme clusters, case mapping and normalization (SPEC section 6), from the Unicode data tables, generated at build time.

## Notes

Outline. Decide when refining whether the tables are generated Narra source or C data.
