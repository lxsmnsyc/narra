---
id: c-abi-export
title: Shared library output with stable C ABI exports
status: todo
priority: could
type: feature
story:
early: false
outline: true
milestone: p5-distribution
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-06
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: c-interop
    kind: hard
    why: builds on `@export(C)`, `@repr(C)`, `Ptr<T>` and `CString` from c-interop
  - on: cross-targets
    kind: hard
    why: a library is built per target, through the same target and C compiler handling
---

# Shared library output with stable C ABI exports

## Goal

`narra build --emit=shared-lib` (and `static-lib`) builds a `.so`, `.dylib`, `.dll` or wasm module whose `@export(C)` functions have a stable C ABI, like Rust's `cdylib` with `extern "C"`. It also writes a C header for the exports. A C test program links the library, calls the exports and runs clean under AddressSanitizer and the leak check.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.

- Gate: the ownership rule for values that cross the boundary needs a decision record before the code. It covers strings and byte buffers passed in (borrowed for the call, or copied) and returned (the library allocates, and the host frees through an exported `narra_free`-style function), plus what reference-counted values may appear in an exported signature at all. A host must never free memory it cannot free, and nothing may leak.
- Runtime state a library needs that a binary does not: initialization without `main`, thread-local state per calling thread, and panics that must not unwind into the host (abort, or return an error code).
- Export names are not hashed (see c-interop), so they need the clash check against the name table.
