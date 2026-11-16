---
id: git-member-deps
title: "Git dependencies on a member of another repository's workspace"
status: active
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
  - roadmaps/selfhost/nodes/git-member-deps/**
  - testdata/deps/workspace_*/**
shared:
  - selfhost/src/driver/depresolve.nra
  - selfhost/src/driver/depcache.nra
  - selfhost/src/driver/modules.nra
  - selfhost/src/driver/diagnostics.nra
  - selfhost/tests/src/gitfixtures.nra
  - SPEC.md
reserves: [diag E0486-E0487]
provides: []
depends_on:
  - on: workspace-manifest
    kind: hard
    why: it reads `[workspace]` in the checkout
  - on: deps-fetch
    kind: behavior
    via: dependencies
  - on: deps-resolve
    kind: behavior
    via: dependencies
  - on: package-dependencies
    kind: hard
    why: it loads the found member's `src/`
---
# Git dependencies on a member of another repository's workspace

## Goal

A git dependency finds its package in the checkout: the root package, or else the member with that name. Tags are `<name>/v<version>` in a workspace repository. A `workspace = true` dependency inside a fetched checkout becomes a git requirement on the same URL, at the sibling's `[package] version`.

## Context

- Design: D101, section "Git dependency on a member of another repository".
- deps-fetch (as amended from D101) takes the tag and lists `<name>/v...` tags first.

## Acceptance

- [ ] A fixture repository with members `json` and `core` is tagged `json/v1.1.0` and `core/v1.0.0`, and `json` has `core = { workspace = true }`. An app with `json = { git, version = "1.1.0" }` builds with `core` 1.0.0 in its lock.
- [ ] The same app also requiring `core` 1.2.0 selects 1.2.0.
- [ ] Errors in E0486-E0487, each with a case. First, no package of that name in the checkout. Second, a sibling with no `version`, or no tag for that version.
- [ ] The cache check is keyed by the tag `listVersions` chooses. Today a version cached under its plain tag `v1.2.0` is used without git even when the repository also has the member tag `json/v1.2.0`, while a clean machine picks the member tag, so two machines can lock different commits for one manifest (inbox, 2026-11-15, from deps-resolve). A case with both tags fails on the base and passes here.
- [ ] SPEC section 15 documents member tags.

## Sizing

M: a package lookup in the checkout and an edge rewrite in the resolver.
