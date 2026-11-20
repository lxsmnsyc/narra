---
id: lsp-dependency-definition
title: "Go to definition into git dependencies"
status: done
priority: could
type: fix
story: batch-20
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
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/lsp-dependency-definition/**
shared:
  - selfhost/src/tools/query.nra
  - selfhost/src/tools/lspnavigation.nra
  - testdata/lsp/**
  - testdata/workspaces/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on: []
---
# Go to definition into git dependencies

## Goal

Go to definition into a git dependency package returns null, because `buildViews` in `selfhost/src/tools/query.nra` (line 835 on main) joins the display file to the package directory. workspace-lsp mapped member files through the `overrides` map in `lspnavigation.nra`, so workspace members work. Fix it in `buildViews` so every dependency kind resolves (git cache, path, member), and drop the member special case if the fix covers it. `narrac query definition` should agree with the LSP.

## Acceptance

- [ ] A case with a git dependency (use the local git fixtures the deps tests use) that returns null on the batch base and the dependency's file and range on the branch.
- [ ] `sh scripts/test.sh` passes.
