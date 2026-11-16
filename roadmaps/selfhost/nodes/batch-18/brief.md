---
id: batch-18
title: "Batch 18: E0971 renames and the workspace manifest"
status: active
priority: should
type: story
story: 
early: false
outline: false
milestone: p3-review
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/batch-18/**
shared:

reserves: []
provides: []
depends_on:
---
# Batch 18: E0971 renames and the workspace manifest

## Goal

Finish p3-review with D124 (restrict-renames), and start p4-packages with the first workspace node (workspace-manifest). The two touch different parts of the compiler and run in parallel.

The branch starts from `main` after PR #122 (batch-17) merged.

## Children

- restrict-renames: E0971 for a second name of a tracked value, its fix, the compiler's own sites, SPEC and guide text.
- workspace-manifest: the `[workspace]` table, `workspace = true` dependencies, `findWorkspace` and `listMembers`.

## Acceptance

- [ ] Every child done or out of the batch with a reason.
- [ ] `sh scripts/test.sh` passes; the seed is current; `pnpm --dir website build` passes.
