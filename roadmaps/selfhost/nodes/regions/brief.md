---
id: regions
title: Regions
status: todo
priority: must
type: feature
outline: true
milestone: p1-syntax
target:
from:
size: M
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
  - on: regions-design
    kind: hard
    why: D100 decides the design and splits this node into children
  - on: shared-write-check
    kind: hard
    why: in-place writes on region values need the D94 shared-write analysis
---

# Regions

## Goal

`Region.run` with bump allocation, copying values out, and no reference count updates inside.

## Notes

Parked at could by the maintainer (2026-11-10) after D100. regions-design found no compiler phase where a region pays off: trees live for the whole compile, `formatSource` per file gains at most 5% over free lists, and the whole compile as one region gains at most 31% at 6 times the peak memory, only after D94. When refined, split into the three children proposed in `nodes/regions-design/outcome.md`: region-runtime, region-copy-out and region-check (two diagnostic codes to reserve then). The SPEC example `Region.run((region) => ...)` changes to `Region.run(() => ...)` when built.
