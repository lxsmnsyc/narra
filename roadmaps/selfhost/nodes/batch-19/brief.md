---
id: batch-19
title: "Batch 19: workspaces"
status: active
priority: could
type: story
story: 
early: false
outline: false
milestone: p4-packages
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
  - roadmaps/selfhost/nodes/batch-19/**
shared:

reserves: []
provides: []
depends_on:
  - on: batch-18
    kind: hard
    why: workspace-manifest landed there
---
# Batch 19: workspaces

## Goal

Finish p4-packages, the lowest open phase: the rest of the D101 workspace design after workspace-manifest (batch-18). At the end this repository is a workspace (maintainer request, 2026-11-14).

The branch starts from `main` after PR #123 (batch-18) merged.

## Children

- workspace-loader: member dependencies, resolve from every member, one lock at the root.
- git-member-deps: git dependencies on a member of another repository's workspace, plus the cache tag fix from the inbox.
- workspace-commands: build, run, test, check, fmt and query at the workspace root.
- workspace-lsp: the language service over members.
- repo-workspace: this repository becomes a workspace. Last.

When every child is done, the umbrella node `workspaces` is set `done` with `fulfilled_by: batch-19`.

## Acceptance

- [x] Every child done or out of the batch with a reason.
- [x] `sh scripts/test.sh` passes; the seed is current; `pnpm --dir website build` passes. test.sh: PASS, 10443 cases on the final tree; bootstrap fixed point, seed current; website build passes; linux-check: PASS before the review fixes.
