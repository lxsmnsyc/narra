---
id: batch-26
title: "Batch 26: yaml, changesets, package C sources"
status: done
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
fulfilled_by: PR #158
gate: false
added: 2026-11-26
spent:
owns:
  - roadmaps/selfhost/nodes/batch-26/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Batch 26: yaml, changesets, package C sources

## Goal

Maintainer, 2026-11-26: after websocket, do yaml, then changesets, then package C sources, in that order.

The branch `feat/batch-26` starts from `feat/batch-25` (PR #157, not merged yet), because the children follow the package pattern batch 25 set up (`packages/*` checks and `examples/*` builds in `scripts/test.sh`). Rebase onto `main` once #157 merges. Worktree `../narra-batch-26`, local only until the story pull request opens.

## Children

- pkg-yaml: first.
- fix-async-arrow-capture: beside pkg-yaml (no shared files), a compiler fix found by the websocket work.
- pkg-changesets: after pkg-yaml.
- package-c-sources: after pkg-changesets. Design approved 2026-11-26: `extern {} from "./file.c"`, a `[c]` table.
- c-packages: after package-c-sources. C libraries as packages that fetch upstream source by git and tag (maintainer, 2026-11-26).

## Acceptance

- [x] Every child is done.
- [x] `sh scripts/test.sh` passes on the branch; `sh selfhost/bootstrap.sh` reaches the fixed point.
