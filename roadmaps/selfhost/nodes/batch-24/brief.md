---
id: batch-24
title: "Batch 24: std modules for servers and release tools"
status: active
priority: should
type: story
story:
early: false
outline: false
milestone: p2-std
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
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/batch-24/**
shared: []
reserves: []
provides: []
depends_on: []
---
# Batch 24: std modules for servers and release tools

## Goal

Maintainer, 2026-11-25: "one shot the p2-std". Every p2-std node added on 2026-11-25 lands in one batch: std-url, std-hex-base64, std-digest, std-log, std-signal, std-semver and process-run-options.

The branch `feat/batch-24` starts from `main` at aa4e6bba. Worktree `../narra-batch-24`, local only until the story pull request opens.

## Children

- std-url, std-hex-base64, std-semver, std-log, std-signal, process-run-options: independent.
- std-digest: after std-hex-base64.

The std nodes share the generated `selfhost/src/std/stdtext.nra` and `docs/reference/std/README.md`. Each child regenerates them on landing instead of merging by hand.

## Acceptance

- [ ] Every child is done.
- [ ] `sh scripts/test.sh` passes on the branch; `sh selfhost/bootstrap.sh` reaches the fixed point.
