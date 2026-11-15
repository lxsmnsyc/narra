---
id: c-abi-export
title: "Shared and static library output with a C header"
status: todo
priority: could
type: feature
story:
early: false
outline: false
milestone: p5-distribution
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
added: 2026-11-06
spent:
owns:
  - roadmaps/selfhost/nodes/c-abi-export/**
  - selfhost/src/cheader.nra
  - testdata/library/**
shared:
  - selfhost/src/driver.nra
  - selfhost/src/manifest.nra
  - selfhost/src/emit.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/*.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1630-E1639]
provides: []
depends_on:
  - on: c-export-functions
    kind: hard
    why: a library is its exported functions; the output needs them to exist
  - on: target-table
    kind: contract
    via: targets
---
# Shared and static library output with a C header

## Goal

`narrac build --emit shared-lib` and `--emit static-lib` (and a `[[library]]` table in `narra.toml` if D122 says so) build a `.dylib`, `.so`, `.dll` or `.a` whose only visible symbols are the `@export(C)` functions, plus a C header declaring them and the C-layout types they use. Like Rust's `cdylib` with `extern "C"`. Per target through `--target` (contract `targets`).

## Context

The old outline depended on cross-targets (the run script); only the target table is needed, so the edge is a contract on target-table. A wasm reactor module, if D122 keeps it, is built and run through wasi-runtime's flags.

## Acceptance

- [ ] A C test program links the shared library and the static library, calls each export, and runs clean under AddressSanitizer and the leak check, in the test harness with `cc` alone.
- [ ] `nm` shows only exported symbols (plus the runtime ones D122 names) in the shared library.
- [ ] The header compiles with `cc -std=c11 -Wall -Werror`.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: driver flags, a header writer, visibility.
