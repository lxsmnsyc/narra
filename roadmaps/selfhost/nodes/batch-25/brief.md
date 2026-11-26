---
id: batch-25
title: "Batch 25: the http package"
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
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent:
owns:
  - roadmaps/selfhost/nodes/batch-25/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 25: the http package

## Goal

Maintainer, 2026-11-26: start the http package with the API approved in the session (recorded in the pkg-http brief). It is the first first-party package under `packages/`.

The branch `feat/batch-25` starts from `main` at 24a59973. Worktree `../narra-batch-25`, local only until the story pull request opens.

## Children

- pkg-http.

## Acceptance

- [ ] Every child is done.
- [ ] `sh scripts/test.sh` passes on the branch; `sh selfhost/bootstrap.sh` reaches the fixed point.
