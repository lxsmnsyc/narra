---
id: std-unicode
title: std/unicode
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

# std/unicode

## Goal

Grapheme clusters, case mapping and normalization (SPEC section 6), from the Unicode data tables, generated at build time.

## Notes

Outline. Decide when refining whether the tables are generated Narra source or C data.
