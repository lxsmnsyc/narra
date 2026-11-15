---
id: deps-fetch
title: "Fetch a dependency version from git into the cache"
status: todo
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
  - roadmaps/selfhost/nodes/deps-fetch/**
  - selfhost/src/depcache.nra
  - selfhost/tests/src/gitfixtures.nra
  - testdata/deps/**
shared:
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/main.nra
  - selfhost/tests/src/registry.nra
reserves: [diag E0440-E0449]
provides: []
depends_on:
  - on: deps-manifest-lock
    kind: contract
    via: dependencies
---

# Fetch a dependency version from git into the cache

## Goal

Given a git URL and a version, put a read-only checkout of tag `v<version>` in the cache and return its commit hash and tree hash. Also list the version tags of a repository, which node deps-resolve needs. All git work goes through `Process.run("git", ...)`, as the driver already runs `cc`.

## Context

- Cache layout and the content hash are fixed in `contracts/dependencies.md`: `$NARRA_CACHE`, else `$HOME/.cache/narra`; content hash is the tree hash.
- Write the checkout to a temporary folder in the cache and rename it into place, so an interrupted fetch never leaves a half-written version.
- A cached version is reused without running git. A cached version whose tree hash no longer matches is an error, not a refetch.
- Tests must not use the network. Add a test runner helper (`gitfixtures.nra`) that builds fixture repositories in the scratch folder with `git init`, commits and tags, and points `NARRA_CACHE` at scratch. Fixture sources live in `testdata/deps/`. Skip these cases with a clear message when `git` is not on `PATH`; `scripts/linux-check.sh` must have git in its image (check it).

## Acceptance

- [ ] `fetchVersion(git, version)` returns the checkout path, commit and tree hash; a second call runs no git command.
- [ ] `listVersions(git)` returns the `vMAJOR.MINOR.PATCH` tags as versions, sorted, ignoring other tags.
- [ ] Errors in E0440-E0449: git not found, repository unreachable, tag missing, cached tree hash mismatch. Each has a case.
- [ ] No test reaches the network; the cases pass with networking off.

## Sizing

M: one module of process calls and file moves, plus a test fixture helper.
