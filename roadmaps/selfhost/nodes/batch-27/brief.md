---
id: batch-27
title: "Batch 27: http bodies and forms at Fetch parity"
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
size: M
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
  - roadmaps/selfhost/nodes/batch-27/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 27: http bodies and forms at Fetch parity

## Goal

Maintainer, 2026-11-26: the body readers and forms still ask every caller for a byte limit, and a program cannot send a form at all. Both were missed when bodies moved to the Fetch model.

The branch `feat/batch-27` starts from `main` at 4fdc5f2f. Worktree `../narra-batch-27`, local only until the story pull request opens.

## Children

- pkg-http-fetch-parity.

## Acceptance

- [ ] Every child is done.
- [ ] `sh scripts/test.sh` passes on the branch.
