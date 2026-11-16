---
id: workspace-lsp
title: "The language service checks a member and the members that depend on it"
status: done
priority: could
type: feature
story: batch-19
early: false
outline: false
milestone: p4-packages
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
added: 2026-11-15
spent:
owns:
  - roadmaps/selfhost/nodes/workspace-lsp/**
  - testdata/lsp/workspace_*/**
shared:
  - selfhost/src/tools/lspserver.nra
  - selfhost/src/tools/lspcache.nra
  - selfhost/src/tools/lspnavigation.nra
  - selfhost/tests/src/lsp.nra
reserves: []
provides: []
depends_on:
  - on: workspace-loader
    kind: hard
    why: the service loads members the way the driver does
---
# The language service checks a member and the members that depend on it

## Goal

`narrac lsp` finds the package from the nearest manifest with `[package]`, then finds its workspace. After each change, it checks that member and its dependent members, each as its own program. Diagnostics go only to each program's own modules.

## Acceptance

- [ ] `findPackageDirectory` skips a manifest with no `[package]`. A single file under such a root is checked alone, as today.
- [ ] An edit in member `json` publishes diagnostics for `json` and for member `app` that imports it. `app`'s publish holds no entry for `json`'s files.
- [ ] An edit in `app` does not check `json`'s bodies again. The `lspcache` per-module state is keyed by package directory.
- [ ] Go to definition from `app` into `json` opens the file in the member directory.

## Sizing

M: package lookup and a per-package cache key in the language service.
