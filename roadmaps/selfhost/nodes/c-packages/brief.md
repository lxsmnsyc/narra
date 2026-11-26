---
id: c-packages
title: "C packages: C libraries fetched from upstream as dependencies"
status: active
priority: could
type: feature
story: batch-26
early: false
outline: false
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent:
owns:
  - roadmaps/selfhost/nodes/c-packages/**
  - testdata/packages/c-packages/**
shared:
  - selfhost/src/driver/*.nra
  - selfhost/src/check/*.nra
  - selfhost/src/driver/diagnostics.nra
  - selfhost/tests/src/*.nra
  - testdata/errors/**
  - docs/guide/**
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: package-c-sources
    kind: hard
    why: extends the [c] table, the C compile step and the object cache
---

# C packages: C libraries fetched from upstream as dependencies

## Goal

A C library is a package like any other: a small Narra package whose manifest points at the upstream source by git URL and tag and says how to build it, with no copy of the source in its repo. Other packages depend on it in `[dependencies]`; version selection, the lock and the cache work as for Narra packages.

## Design (maintainer approved, 2026-11-26)

```toml
[package]
name = "c_zlib"
version = "1.3.1"

[c]
git = "https://github.com/madler/zlib"
tag = "v1.3.1"
sources = ["*.c", "!example.c", "!minigzip.c"]
include = ["."]
```

- `[c] git` and `tag` make the `[c]` paths relative to the fetched upstream checkout instead of the package folder. `narrac` fetches the tag into the dependency cache like a git dependency, and `narra.lock` records its commit and tree hash, so every machine builds the same bytes. A changed tree for a locked tag is an error, as for git dependencies.
- The package version follows upstream; an update is a version and tag change, released with `changesets`.
- A dependent binds it with `extern { ... } from "c_zlib/zlib.h"`: `<package>/<path>` names a file of that C package's source and links that package's objects. The dependent gets the package's `include` folders; never its `defines`.
- One copy per program: version selection picks one version of a C package, compiled once and linked once.
- A fetch needs the network only when the cache lacks the tag, as for git dependencies; the language service never fetches.

## Context

- Read `roadmaps/selfhost/nodes/package-c-sources/outcome.md` first: the `[c]` table, `from "./path"`, the compile step and the object cache.
- The dependency fetcher and lock live in `selfhost/src/driver/` (`depcache.nra`, `depresolve.nra`, `depload.nra`, the lockfile module). Reuse them; the upstream checkout is a lock entry of its own kind or a field of the C package's entry (the worker decides and writes it in SPEC section 15).
- Tests must not need the network: build the upstream as a local git repo in the test harness (a `file://` URL), with a tag.

## Acceptance

- [ ] A test C package points at a local git repo of a small C library (a few `.c` files and a header) by `file://` URL and tag; a Narra package depends on it, binds a function with `from "<package>/<header>"`, and calls it, clean under AddressSanitizer and the leak check.
- [ ] `narra.lock` records the upstream commit and tree; a second build fetches nothing; a moved tag (a different tree) is an error naming the package.
- [ ] Two packages depending on the C package at different versions build one copy (checked through the link or the cache).
- [ ] A missing tag, a missing header in `from`, and `defines` leaking to a dependent are each tested (the last must not happen).
- [ ] SPEC section 15 and the C interop guide describe C packages.
- [ ] `sh scripts/test.sh` passes and `sh selfhost/bootstrap.sh` reaches the fixed point.

## Sizing

M: a fetch kind on the existing fetcher, a lock field, `from "<package>/<path>"` resolution, include folders passed to dependents.
