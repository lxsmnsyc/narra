---
id: cross-targets
title: Cross-compilation targets, including wasm32-wasi
status: todo
priority: could
type: feature
story:
early: false
outline: true
milestone: m13-distribution
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-06
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: target selection is built once, in the self-hosted driver, not in the Go compiler
---

# Cross-compilation targets, including wasm32-wasi

## Goal

`narra build --target <triple>` builds a binary for another platform from one host: darwin and linux on arm64 and x64, linux musl, windows x64, and `wasm32-wasi`. The triple passes through to the C compiler (`zig cc -target <triple>`, or clang with `--target` and a sysroot; wasi-sdk for wasm). The runtime in `runtimetext.nra` compiles and passes the sample programs on each target. A CI-free script builds every target from one machine, and the samples run under wasmtime for `wasm32-wasi`.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.

- Today `selfhost/src` only emits C. The Go driver (`compiler/internal/narra/driver.go`) and `selfhost/bootstrap.sh` call `cc` with fixed flags (`-std=c11 -w ... -lm`). Decide where `narra build` lives in the self-hosted compiler first, together with c-seed if that has landed.
- Runtime portability to check: `_Thread_local`, `getenv`, `fopen` and the file APIs used by std, `-lm`, and anything POSIX-only. wasm32-wasi has no threads by default, so threads and async must be gated or stubbed per target.
- build-conditions should read the same target description for `std/build` constants and `@target`, so the two nodes share one target table. Whichever lands first defines it.
- Source: the 2026-11-06 discussion on shipping Narra tools to npm as per-platform packages, the way esbuild, swc and oxc do.
