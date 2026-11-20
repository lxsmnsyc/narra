---
id: batch-21
title: "Batch 21: docs that teach every crucial concept"
status: done
priority: could
type: story
story:
early: false
outline: false
milestone: p6-docs
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/batch-21/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-20
    kind: hard
    why: starts from main after PR #125
---
# Batch 21: docs that teach every crucial concept

## Goal

Maintainer, 2026-11-20: the docs are missing a lot of crucial concepts. The first release (`v0.1.0`) waits for this batch.

The branch `feat/batch-21` starts from `main` after PR #125 (batch-20). Worktree `../narra-batch-21`, local only.

## Children

- docs-concept-audit: list every concept a user needs, map it to the docs, and propose what to write. Gate: the maintainer approves the list.
- Writing nodes from the approved audit.

## Acceptance

- [x] The audit approved; every concept it marks crucial is taught in the guide or documented in a reference page.
- [x] `pnpm --dir website build` passes; the guide suite passes. test.sh: PASS, 10820 cases; bootstrap fixed point, seed current.
