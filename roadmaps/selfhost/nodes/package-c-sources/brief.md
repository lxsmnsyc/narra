---
id: package-c-sources
title: "Packages compile their own C sources"
status: todo
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
added: 2026-11-25
spent:
owns:
  - roadmaps/selfhost/nodes/package-c-sources/**
  - testdata/packages/c-sources/**
shared:
  - selfhost/src/driver/*.nra
  - selfhost/src/diagnostics*.nra
  - selfhost/tests/src/*.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1660-E1662]
provides: []
depends_on: []
---

# Packages compile their own C sources

## Goal

A package lists C source files in its `narra.toml`, and `narrac build` compiles them with the same `cc`, target and flags as the emitted C, then links them into the program. This lets a package bind a C library it ships (Lua, zlib, SQLite) and write C wrappers for variadic functions.

## Context

- Source: `inbox.md` 2026-11-20 (guide-c-interop): "A package cannot compile C sources yet". `GOAL.md` Decisions, 2026-11-21: the C-backed packages (`tls`, `crypto`, `sqlite`, `compress`, `intl`) wait for this node. SPEC section 9 names the variadic gap ("waits for C-callable functions and C sources in a package").
- `[link]` today has one key, `search` (SPEC section 15); any other key is E1000. Add a key for sources (and include folders and defines), relative to the package directory. Decide the names in the node and write them in SPEC.
- Build per target through `--target`, including `Wasi-Wasm32` with the flags of the wasm row. Cache object files by content hash under the existing build cache so a rebuild does not recompile unchanged C.
- Callbacks from C into Narra are node `c-export-functions`, not this one.

## Acceptance

- [ ] A test package ships a `.c` file with a variadic wrapper (`snprintf` behind a fixed signature) and a Narra sample in another package depends on it and calls it, clean under AddressSanitizer and the leak check.
- [ ] A missing source file, a source outside the package directory and a C compile error each give their own diagnostic (E1660-E1662) naming the package and file.
- [ ] A second build with no change compiles no C file (checked by the build log or timing in the test harness).
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: one manifest key, a compile step in the driver, an object cache.
