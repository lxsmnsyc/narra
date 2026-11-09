---
id: package-dependencies
title: "Package dependencies"
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
size: L
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
  - on: batch-2
    kind: hard
    why: builds on the current manifest and module loader
---

# Package dependencies

## Goal

SPEC.md section 15: `[dependencies]` from git repositories pinned to version tags, minimal version selection, one version per package, `narra.lock` with commit and content hashes. Today `[dependencies]` is E0407. Split when refined (fetch and cache, resolution, lock file, loader).

## Notes

Outline. Refine from the decision record named above.
