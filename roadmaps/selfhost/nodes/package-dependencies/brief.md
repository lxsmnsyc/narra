---
id: package-dependencies
title: "Package dependencies: load dependency modules in build, check, test and the language service"
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
added: 2026-11-09
spent:
owns:
  - roadmaps/selfhost/nodes/package-dependencies/**
  - testdata/packages/deps_*/**
shared:
  - selfhost/src/modules.nra
  - selfhost/src/hash.nra
  - selfhost/src/driver.nra
  - selfhost/src/lspserver.nra
  - selfhost/src/scopes.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/packages.nra
  - SPEC.md
  - docs/guide/*.md
reserves: [diag E0460-E0469]
provides: []
depends_on:
  - on: deps-resolve
    kind: hard
    why: the loader reads the resolved versions and their checkout paths; there is nothing to load before resolution exists
  - on: deps-fetch
    kind: behavior
    via: dependencies
---

# Package dependencies: load dependency modules

## Goal

The last step of SPEC section 15 dependencies. After resolution, the loader loads each dependency package's `src/` from its cache checkout, so `import * as Json from "json/parse"` works in `narra build`, `run`, `test`, `check` and the language service. This node was an outline for all of package dependencies; deps-manifest-lock, deps-fetch and deps-resolve now hold the other parts.

## Context

- `loadPackage` in `selfhost/src/modules.nra` loads one package's `src/` recursively, then std, then orders modules. Add the dependency packages between the two, each module path prefixed by its package name.
- `moduleHash` in `hash.nra` hashes the module path without the package name, so two packages that both have `src/index.nra` get the same C names (E0417). Include the package name for dependency packages only, so the root package's C names, every golden file and the seed stay the same.
- A dependency's modules may import only their own package, its declared dependencies and std. Importing a package that is not a declared dependency is an error.
- Tests in dependency packages do not run under `narra test` of the root package.
- The language service finds the package from the nearest `narra.toml` (`lspserver.nra` line 213). It reads the lock and the cache and never fetches; a missing checkout is a diagnostic that says to run `narra build`.

## Acceptance

- [ ] A fixture app with two git dependencies (one shared, from the deps-resolve diamond) builds and runs under the sanitizers and leak check.
- [ ] Two packages with the same module file names link without a hash clash; the root package's emitted C is unchanged (golden files untouched).
- [ ] Errors in E0460-E0469: importing an undeclared package, a missing checkout in the language service, an import cycle across packages.
- [ ] SPEC section 15 and the guide's packages chapter describe dependencies as built; E0407 is gone everywhere.

## Sizing

M: loader and driver wiring on top of the three earlier nodes.
