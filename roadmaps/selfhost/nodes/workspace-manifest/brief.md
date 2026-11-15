---
id: workspace-manifest
title: "[workspace] table, workspace = true dependencies and finding the workspace root"
status: todo
priority: could
type: feature
story:
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
  - roadmaps/selfhost/nodes/workspace-manifest/**
  - selfhost/src/workspace.nra
  - testdata/workspaces/manifest_*/**
  - testdata/errors/workspace_*
shared:
  - selfhost/src/manifest.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/manifest.nra
  - selfhost/tests/src/packages.nra
reserves: [diag E0480-E0483]
provides: []
depends_on:
  - on: deps-manifest-lock
    kind: contract
    via: dependencies
---
# `[workspace]` table, `workspace = true` dependencies and finding the workspace root

## Goal

`parseManifest` reads `[workspace] members` and `json = { workspace = true }`. A new module `selfhost/src/workspace.nra` exports two functions. `findWorkspace(packageDirectory): Option<Workspace>` finds the workspace. `listMembers(root): Result<Array<Member>, Array<LoadError>>` lists each member's directory and name. Nothing loads members yet.

## Context

- Design: D101 in `DECISIONS.md`, from node workspace-design.
- Builds against `contracts/dependencies.md` version 2, where `Dependency` has `workspace: bool`.

## Acceptance

- [ ] `members` takes exact paths, `*` as one whole segment, and `!` entries that remove matches. Members come out sorted by directory.
- [ ] Each of these errors has a case and a code in E0480-E0483:
  - `members` is not an array of strings.
  - A member has no `narra.toml`, or has its own `[workspace]`.
  - Two members have one name.
  - A `workspace = true` dependency has another key beside it.
- [ ] `findWorkspace` returns the first `[workspace]` above a package only when its `members` list the package. An unlisted package gets `None`.
- [ ] A root with `[package]` and `[workspace]` lists itself as a member.

## Sizing

M: one parser addition and one module of path matching.
