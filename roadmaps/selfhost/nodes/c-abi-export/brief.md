---
id: c-abi-export
title: "Shared and static library output with a C header"
status: todo
priority: later
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
reserves: [diag E1630-E1632]
provides: []
depends_on:
  - on: c-export-functions
    kind: hard
    why: a library is its exported functions; the output needs them to exist; D123 builds on it
  - on: target-table
    kind: contract
    via: targets
  - on: wasi-runtime
    kind: behavior
    via: targets
---
# Shared and static library output with a C header

## Goal

`narrac build --emit shared-lib` and `--emit static-lib` (flags only, no `[[library]]` table; the name must be a C identifier, E1630; file names are in D122 item 5) build a `.dylib`, `.so`, `.dll` or `.a` whose only visible symbols are the `@export(C)` functions, plus a C header declaring them and the types they use. Like Rust's `cdylib` with `extern "C"`. Per target through `--target` (contract `targets`). `--emit shared-lib --target Wasi-Wasm32` writes a reactor module whose only exports are the `@export(C)` functions plus the allocator calls D122 names, with no `main`.

## Context

The old outline depended on cross-targets (the run script); only the target table is needed, so the edge is a contract on target-table. The wasm reactor module is a required output (D123 item 2): it is built and run through wasi-runtime's flags.

The header has `narra_buffer`, status macros, one prototype per entry, the three support functions, `NARRA_API` and a C++ guard. Codes: E1630-E1632.

## Acceptance

- [ ] A C test program links the shared library and the static library, calls each export, and runs clean under AddressSanitizer and the leak check, in the test harness with `cc` alone.
- [ ] `nm` shows only the entries and the support functions, nothing else, in the shared library.
- [ ] A reactor module built with `--target Wasi-Wasm32` runs one export under wasmtime.
- [ ] `static-lib` with `--target Wasi-Wasm32` is E1631. A program with no export is E1632.
- [ ] The reactor test uses the wasm flags of D122 item 5. wasmtime runs it with `-W exceptions=y`.
- [ ] The header compiles with `cc -std=c11 -Wall -Werror`.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: driver flags, a header writer, visibility.
