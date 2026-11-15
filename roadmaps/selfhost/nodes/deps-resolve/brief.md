---
id: deps-resolve
title: "Minimal version selection and writing narra.lock"
status: done
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
  - roadmaps/selfhost/nodes/deps-resolve/**
  - selfhost/src/depresolve.nra
  - testdata/deps_resolve*
shared:
  - selfhost/src/diagnostics.nra
  - selfhost/src/manifest.nra
  - selfhost/tests/src/packages.nra
  - testdata/deps/**
reserves: [diag E0450-E0459]
provides: []
depends_on:
  - on: deps-manifest-lock
    kind: contract
    via: dependencies
  - on: deps-fetch
    kind: behavior
    via: dependencies
---

# Minimal version selection and writing narra.lock

## Goal

SPEC section 15: walk the dependency graph from the root manifests, reading each dependency's `narra.toml` at its required version, and select for every package the highest version some package in the graph requires as its minimum. One version per package. The result is checked against `narra.lock`, or written to it when the lock is missing or a dependency was added.

## Context

- Build against a fake fetch (a map from URL and version to a manifest text) first; switch to `fetchVersion` from deps-fetch when it lands.
- Minimal version selection needs no solver: repeat "raise each package to the highest minimum seen" until nothing changes. Only versions named in some manifest are fetched, not every tag.
- **Several roots (D101).** The resolver takes an array of root manifests. A standalone package passes one. workspace-loader passes every member. No workspace knowledge lives here.
- **`[overrides]` (D101).** The root of the build may have `[overrides]`, with entries shaped like a `Dependency`. Every requirement on an overridden name uses the override's URL, and the selected version is the override's version exactly, even below a minimum. An overridden name reached through two URLs is not an error, because the override replaces both. The lock records the selected URL and version.
- `parseManifest` reads `[overrides]`. If deps-manifest-lock has not added the table, add it here in `manifest.nra`.
- A package name reached through two different git URLs, with no override, is an error. A major version change is a new package name, so two majors of one name never meet.
- `narra build` fails when the lock's commit or tree hash differs from the fetched one (SPEC). A lock entry no manifest needs is dropped on the next write.

## Acceptance

- [ ] Fixture graphs in `testdata/deps/`: a diamond where two packages require different minimums of a third, and the higher one wins; a chain three deep; a cycle between two packages (allowed by selection, reported later by the loader if modules import in a cycle).
- [ ] Errors in E0450-E0459: one name with two URLs, a required version with no tag, a lock hash mismatch, a dependency whose manifest names a different package. Override errors take codes from the same range: `[overrides]` in a manifest that is not the root of the build, an override for a name no package requires, and an override whose tag is missing.
- [ ] A diamond fixture where the override pins the shared package to a fork URL and a lower version, and the lock shows the fork. An `[overrides]` table in a dependency's manifest is an error.
- [ ] Two root manifests that require different minimums of one package select the higher.
- [ ] The lock written for a fixture is byte-identical on a second run.

## Sizing

M: one resolver module over the contract types, tested on fixtures.
