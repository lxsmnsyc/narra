---
id: narrac-wasm
title: "narrac built for Wasi-Wasm32"
status: todo
priority: could
type: feature
story: batch-16
early: false
outline: false
milestone: p5-distribution
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/wasm-build.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-15
spent:
owns:
  - roadmaps/selfhost/nodes/narrac-wasm/**
  - scripts/wasm-build.sh
shared:
  - scripts/cross-check.sh
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: wasi-runtime
    kind: hard
    why: the runtime fixes for `Env.args` and `/dev/stdout` that narrac needs on wasi
  - on: target-table
    kind: contract
    via: targets
---
# narrac built for Wasi-Wasm32

## Goal

`sh scripts/wasm-build.sh` builds the compiler from the seed for `Wasi-Wasm32`, strips it with `llvm-strip`, and writes `dist/narrac.wasm`. cross-targets runs `check` and `emit` with it on its sample set and compares the result with native.

## Context

- Source: D123 item 8 and the T1 proof in `debate/js-interop/evidence/T1-narrac-wasm.md`. The spike needed three patches to the runtime and `-lsetjmp`; wasi-runtime carries the fixes.
- Do not use `wasm-opt --strip-debug --all-features`: wasmtime 50 refuses its output. Use `llvm-strip`.
- The build takes 68 s at low load and 190 s at load 10 to 15.

## Acceptance

- [ ] Under wasmtime, `emit` gives C byte-identical to native for `testdata/hello.nra`, `regex.nra` and `selfhost/`.
- [ ] The stripped module stays at most 1.2 MB brotli, and the script prints the size.
- [ ] The module validates with wasm exceptions turned off.
- [ ] The script is not part of `scripts/test.sh`, since it needs zig.

## Sizing

S: one script over flags the spike already found.
