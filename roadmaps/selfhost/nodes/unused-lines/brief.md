---
id: unused-lines
title: "The unused-names check splits sources only when needed"
status: active
priority: should
type: fix
story: batch-3
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - roadmaps/selfhost/nodes/unused-lines/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 compiler
---

# The unused-names check splits sources only when needed

## Goal

Inbox 2026-11-09 (from lsp-incremental): `reportUnusedTopLevel` in `selfhost/src/unused.nra` splits every module source into lines even when nothing is unused, about 7% of a whole `narrac check`. Split only when there is something to report.

Prove it first: profile or time `narrac check` on selfhost/src on the base and the branch, several runs.

## Acceptance

- [ ] Timings on the base and the branch, with the gain beyond the noise (or "Not reproduced" with the numbers).
- [ ] Diagnostics unchanged.
- [ ] `sh scripts/test.sh` passes.
