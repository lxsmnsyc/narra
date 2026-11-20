---
id: error-index
title: "Generated error index"
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
  - roadmaps/selfhost/nodes/error-index/**
  - docs/reference/errors.md
  - scripts/error-index.sh
shared:
  - selfhost/src/driver/**
  - selfhost/src/tools/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on:
  - on: site-reference-section
    kind: hard
    why: the page lives in its section
---
# Generated error index

## Goal

`narrac explain --list --json` and a generated, committed `/reference/errors` (479 codes, 96 with explanations), checked like the std reference. Proposal change 15.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refined 2026-11-20 into batch-21 (docs-concept-audit, approved).

## Acceptance

- [ ] `narrac explain --list --json` lists every code with its message and explanation; `/reference/errors` is generated from it, committed, and a test fails when it is stale.
- [ ] `sh scripts/test.sh` and `pnpm --dir website build` pass.
