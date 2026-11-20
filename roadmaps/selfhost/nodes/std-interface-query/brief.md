---
id: std-interface-query
title: "narrac query interface for every public std module"
status: todo
priority: could
type: feature
story: 
early: false
outline: true
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-interface-query/**
shared:
reserves: []
provides: []
depends_on:
  - on: first-release
    kind: hard
    why: after the release (docs-structure-review proposal, approved 2026-11-20)
---
# narrac query interface for every public std module

## Goal

`narrac query interface std/<module>` answers for every public std module, builtin and source; today it prints `unknown module` for `std/iterator`, `std/string`, `std/toml`, and 18 builtin modules have no source. Proposal change 10.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refine into a full brief when `plan.py refine` lists it.
