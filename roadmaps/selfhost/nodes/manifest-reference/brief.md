---
id: manifest-reference
title: "narra.toml reference"
status: todo
priority: could
type: docs
story: batch-21
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/manifest-reference/**
  - docs/reference/narra-toml.md
shared:
reserves: []
provides: []
depends_on:
  - on: site-reference-section
    kind: hard
    why: the page lives in its section
---
# narra.toml reference

## Goal

`/reference/narra-toml`: every table and key, type, default, and the error code for a wrong value, gathered from chapters 1, 8, 12 and `SPEC.md` sections 2 and 15. Proposal change 14.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refined 2026-11-20 into batch-21 (docs-concept-audit, approved).

## Acceptance

- [ ] `/reference/narra-toml` documents every table and key (`[package]`, `[dependencies]`, `[exports]`, `[link]`, `[fmt]`, workspaces, `[overrides]`, binaries), its type, default and the error code for a wrong value, checked against the loader code.
- [ ] `pnpm --dir website build` passes.
