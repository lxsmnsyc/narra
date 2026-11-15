---
id: repo-workspace
title: "This repository becomes a workspace"
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
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-15
spent:
owns:
  - roadmaps/selfhost/nodes/repo-workspace/**
  - narra.toml
shared:
  - selfhost/CLAUDE.md
  - README.md
  - scripts/test.sh
reserves: []
provides: []
depends_on:
  - on: workspace-commands
    kind: hard
    why: `check` and `fmt` at the root need the commands
  - on: workspace-lsp
    kind: hard
    why: without it, the lsp suite reaches the root manifest from `testdata/lsp`
  - on: selfhost-layout
    kind: soft
    why: both edit `selfhost/CLAUDE.md`, and either order works
---
# This repository becomes a workspace

## Goal

Add a root `narra.toml` with `[workspace] members = ["selfhost", "selfhost/tests"]`, no `[package]` and no `[fmt]`. `narrac check .` and `narrac fmt --check .` at the root cover the compiler and the test runner. `std/`, `bench/` and `testdata/` stay outside. Requested by the maintainer on 2026-11-14.

## Acceptance

- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point and reports the seed current with no seed change. `narrac emit selfhost` output is byte-identical before and after.
- [ ] `sh scripts/test.sh` passes with the case total unchanged. The lsp, fmt and sources suites show no diff.
- [ ] `narrac check .` at the root exits 0 and checks both members.
- [ ] A test in the `packages` suite fails when the root manifest is removed or a member is dropped from it.

## Sizing

S: one manifest and the checks that guard it.
