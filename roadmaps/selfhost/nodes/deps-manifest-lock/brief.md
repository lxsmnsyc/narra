---
id: deps-manifest-lock
title: "[dependencies] in narra.toml and the narra.lock format"
status: active
priority: could
type: feature
story: batch-16
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
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/deps-manifest-lock/**
  - selfhost/src/lockfile.nra
  - testdata/errors/manifest_dependencies*
  - testdata/lockfile*
shared:
  - selfhost/src/manifest.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/manifest.nra
  - SPEC.md
reserves: [diag E0430-E0439]
provides: [dependencies]
depends_on: []
---

# [dependencies] in narra.toml and the narra.lock format

## Goal

Freeze the `dependencies` contract (`contracts/dependencies.md`) by building its data half: `narra.toml` reads `[package] version` and a `[dependencies]` table of inline tables (`json = { git = "...", version = "1.2.0" }`), and a new module `selfhost/src/lockfile.nra` reads and writes `narra.lock`. Nothing fetches or loads a dependency yet; the loader ignores the parsed list until node package-dependencies uses it.

## Context

- `selfhost/src/manifest.nra` reads a TOML subset by hand: values are strings or arrays of strings (`enum Value`). Inline tables are new. Either add an inline table value to this reader, or move the manifest onto `std/toml` if the committed seed compiles it; pick the smaller change and say which in the outcome.
- Today `[dependencies]` returns E0407 at `manifest.nra` line 92. Remove E0407 from the reader, the `codes` table and its tests (no deprecation path).
- E0400-E0499 is the module and manifest family. E0427-E0499 is free; this node takes E0430-E0439.

## Acceptance

- [ ] `Manifest` has `version` and `dependencies: Array<Dependency>` with the fields in the contract.
- [ ] Errors with codes in E0430-E0439: a dependency without `git` or `version`, a version that is not `MAJOR.MINOR.PATCH`, an unknown key in a dependency, a dependency named `std` or the package's own name, a duplicate name. Each has a case in `testdata/errors/`.
- [ ] `parseLock` and `formatLock` round-trip; `formatLock` sorts by name and is stable. Covered by manifest suite cases.
- [ ] SPEC section 15 lists the codes and the lock file shape.
- [ ] Set the contract to `status: frozen`, version 1, in the outcome (the orchestrator commits it).

## Sizing

M: one manifest extension and one small new module, no I/O.
