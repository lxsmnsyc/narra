---
id: workspace-commands
title: "build, run, test, check, fmt and query at the workspace root"
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
  - roadmaps/selfhost/nodes/workspace-commands/**
  - testdata/workspaces/commands_*/**
shared:
  - selfhost/src/driver/driver.nra
  - selfhost/src/tools/fmt.nra
  - selfhost/src/tools/fmtconfig.nra
  - selfhost/src/tools/query.nra
  - selfhost/src/main.nra
  - selfhost/src/driver/diagnostics.nra
  - selfhost/tests/src/packages.nra
  - SPEC.md
  - docs/guide/*.md
reserves: [diag E0488-E0490]
provides: []
depends_on:
  - on: workspace-loader
    kind: hard
    why: the commands act on loaded members
---
# `build`, `run`, `test`, `check`, `fmt` and `query` at the workspace root

## Goal

The driver commands act on every member when given the workspace root, and on one member when given its directory, as in the D101 command table. `-dependents` and `-dependencies` add related members. `fmt` uses the root `[fmt]` for members without one. SPEC section 15 and the guide's packages chapter describe workspaces.

## Acceptance

- [ ] At the root, each command covers every member:
  - `build` makes every binary of every member.
  - `test` runs each member's tests in `members` order.
  - `check` reports every member's diagnostics, sorted as today.
  - `fmt --check` covers every member.
  - `query def NAME` searches every member.
- [ ] `test packages/json -dependents` also runs the tests of `app`. `check packages/app -dependencies` also checks `json`.
- [ ] `run -bin NAME` at the root finds the binary across members. Two members with one binary name is E0488. A flag `-dependents` or `-dependencies` with a path that is not a member is E0489.
- [ ] `findConfig` skips a `narra.toml` with no `[package]` and no `[fmt]`, so files under `testdata/` keep the defaults.
- [ ] SPEC section 15 documents `[workspace]`, `workspace = true`, `[overrides]` and the root lock. D101 drops "Not built yet" for the parts built.

## Sizing

M: command dispatch over members in the driver, plus docs.
