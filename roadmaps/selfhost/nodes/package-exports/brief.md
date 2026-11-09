---
id: package-exports
title: "Package exports table"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: m13-distribution
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: package-dependencies
    kind: hard
    why: exports only matter once other packages can import this one
---

# Package exports table

## Goal

D97: `[exports]` in narra.toml lists the import paths other packages may use; unlisted modules are internal; the bare package name maps to one module; `narra move` updates the table; std uses the same table so `std/internal` becomes an unexported path.

## Notes

Outline. Refine from the decision record named above.
