---
id: std-interface-query
title: "narrac query interface for every public std module"
status: active
priority: could
type: feature
story: batch-21
early: false
outline: false
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
  - testdata/query/**
shared:
  - selfhost/src/tools/**
  - selfhost/src/driver/**
  - selfhost/tests/src/**
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---
# narrac query interface for every public std module

## Goal

`narrac query interface std/<module>` answers for every public std module, builtin and source; today it prints `unknown module` for `std/iterator`, `std/string`, `std/toml`, and 18 builtin modules have no source. Proposal change 10.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refined 2026-11-20 into batch-21 (docs-concept-audit, approved).

## Acceptance

- [ ] `narrac query --json interface std/<module>` answers for every public std module (builtin and source), with each export's signature and effects; a case covers a builtin and a source module.
- [ ] `sh scripts/test.sh` passes.
