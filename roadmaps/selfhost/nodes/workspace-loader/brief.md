---
id: workspace-loader
title: "Load member dependencies, resolve from every member, one lock at the root"
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
  - roadmaps/selfhost/nodes/workspace-loader/**
  - testdata/workspaces/load_*/**
shared:
  - selfhost/src/driver/modules.nra
  - selfhost/src/driver/diagnostics.nra
  - selfhost/tests/src/packages.nra
  - testdata/deps/**
reserves: [diag E0484-E0485]
provides: []
depends_on:
  - on: workspace-manifest
    kind: hard
    why: it needs `findWorkspace` and `listMembers`
  - on: package-dependencies
    kind: hard
    why: the dependency loader and the hash rule
  - on: deps-resolve
    kind: behavior
    via: dependencies
  - on: package-exports
    kind: hard
    why: exports between members are tested here
---
# Load member dependencies, resolve from every member, one lock at the root

## Goal

`Modules.load` on a member finds its workspace and loads each `workspace = true` dependency's `src/` as a dependency package. Each such package is hashed with its name and checked against `[exports]`. The loader passes every member's manifest to deps-resolve (which takes an array of roots) and reads and writes `narra.lock` at the workspace root.

## Context

- Design: D101. Dependency modules hash with a separator no module path holds (`json:parse`), or the clash check runs over every package in the build. Plain `json/parse` collides with the root package's `src/json/parse.nra` (workspace-design, gotchas).

## Acceptance

- [ ] A fixture workspace builds and runs from the `app` directory under the sanitizers and leak check. In it, `app` depends on `json` by `workspace = true`, and `json` depends on the git fixture `http`.
- [ ] Two members require different minimums of one git package. The higher one is selected, whichever member is built. The root lock is byte-identical on a second run.
- [ ] Errors in E0484-E0485, each with a case. First, a `workspace = true` name with no such member, or a dependency outside a workspace. Second, a `narra.lock` inside a member. A git dependency with a member's name uses the deps-resolve error for one name with two sources.
- [ ] A member with `[exports]` exposes only its listed paths to another member.
- [ ] Building a member alone emits the same C for its own modules as before (golden files untouched).

## Sizing

M: loader changes in one shared module, plus fixtures.
