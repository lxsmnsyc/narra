---
id: std-semver
title: "std/semver: parse and compare versions"
status: todo
priority: should
type: feature
story: batch-24
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: S
risk: low
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/std-semver/**
  - std/src/semver.nra
  - testdata/semver*
shared:
  - selfhost/src/std/stdtext.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
reserves: []
provides: []
depends_on: []
---

# std/semver: parse and compare versions

## Goal

A new `std/semver` module parses Semantic Versioning 2.0.0 versions, compares them with prerelease precedence, bumps them (`major`, `minor`, `patch`), and checks a version against the requirement syntax `narra.toml` uses for dependencies.

## Context

- Source: std-gap-survey rank 27 (`driver/depcache.nra` `isVersionBefore`, `driver/manifest.nra` `isVersion`); `GOAL.md` Decisions, 2026-11-21 (semver stays in std).
- Read SPEC section 14 for the requirement syntax `narra.toml` accepts, and implement exactly that, no more. The `pkg-changesets` node uses the bump functions.
- Replacing the compiler's own copies waits for a seed that knows the module.

## Acceptance

- [ ] The precedence example of semver.org section 11 sorts in the stated order.
- [ ] Invalid versions (`1.2`, `01.2.3`, `1.2.3-`) give `None`.
- [ ] `Semver.bump` resets lower parts and drops the prerelease.
- [ ] `sh scripts/test.sh` passes.

## Sizing

S: a parser, a compare function and the requirement check.
