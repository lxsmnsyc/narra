---
id: batch-20
title: "Batch 20: pre-release fixes"
status: active
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
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/batch-20/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-19
    kind: hard
    why: starts from main after PR #124
---
# Batch 20: pre-release fixes

## Goal

Fix what a first user meets on day one, and review the docs, before the first release tag (maintainer, 2026-11-19: release before the C export chain; c-export-functions, c-abi-export, node-addon and wasm-js-glue wait for a user).

The branch `feat/batch-20` starts from `main` after PR #124 (batch-19) merged. Worktree `../narra-batch-20`, local only.

## Children

- narrac-version: `narrac --version`; installers and docs use it.
- const-float-exact: `@const` evaluation reads long float literals.
- effect-export-gap: reproduce the reported `uses io` gap first; fix it if it shows.
- regex-step-limits: regex timing cases stop depending on wall-clock time.
- lsp-dependency-definition: go to definition into git dependencies.
- lsp-hover-parity: LSP hover matches `narrac query hovers`, plus the batch-17 hover nits.
- export-key-diagnostics: diagnostics and dumps show the export key the importer wrote.
- docs-structure-review: proposal for the maintainer (gate).

## Acceptance

- [ ] Every child done or out of the batch with a reason.
- [ ] `sh scripts/test.sh` passes; the seed is current; `pnpm --dir website build` passes.
