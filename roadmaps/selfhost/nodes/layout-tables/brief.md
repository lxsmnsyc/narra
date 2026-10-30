---
id: layout-tables
title: Layout tables for debug builds
status: todo
priority: later
type: feature
outline: true
milestone: m6-performance
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
  - on: traits
    kind: hard
    why: layout tables carry trait implementation tables too
---

# Layout tables for debug builds

## Goal

Compile generic code once with layout tables in debug builds, as `SPEC.md` section 11 says, and specialize in release builds.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
