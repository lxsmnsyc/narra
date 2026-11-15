---
id: c-abi-design
title: "Spike: ownership and runtime rules at the C library boundary"
status: todo
priority: could
type: spike
story:
early: false
outline: false
milestone: p5-distribution
target:
ship: auto
land: auto
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/c-abi-design/**
shared:
  - DECISIONS.md
reserves: [decision D122]
provides: []
depends_on:
  - on: js-interop-design
    kind: hard
    why: held until the design picks Node-API or another use of a C library boundary (maintainer, 2026-11-14); drop this chain if it does not
---
# Spike: ownership and runtime rules at the C library boundary

## Goal

Write decision D122, in the style of D109, for C interop step 2 (SPEC section 9, "Not built yet"): C-callable Narra functions, C layouts, and library output. The maintainer reads it before c-layout, c-export-functions or c-abi-export start (gate).

## Questions

1. **Ownership across the boundary.** Strings and bytes passed in: borrowed for the call, or copied. Returned: the library allocates, the host frees through one exported free function. Which reference-counted types may appear in an exported signature at all, and whether an opaque handle type (`Ptr<T>` to a Narra value, with exported retain and release) is the way to hand out structs that hold strings or arrays. Nothing may leak, and a host must never free memory it cannot free.
2. **Layouts.** SPEC names `@repr(C)` and `@layout(C)` in different places. Pick one spelling for structs and one for enums with fixed integer values, and the field types allowed (numbers, `bool`, `Ptr<T>`, nested C structs, fixed arrays).
3. **Exported functions.** `@export(C, "name")` rules: allowed signatures, no type parameters, effects (does a C caller need `unsafe`?), name clash check against the name table (E1001 rules from c-interop), and whether an exported function may also be called from Narra.
4. **Runtime state in a library.** Initialization without `main` (lazy on first call, or an exported init), per-thread heaps (D111) when an unknown host thread calls in, the async loop, and what a panic does (abort the process, or return an error code through an out parameter). A panic must never unwind into the host.
5. **Output.** `--emit shared-lib` and `static-lib` names per target (`.dylib`, `.so`, `.dll` with an import lib, `.a`, wasm reactor module), the generated header, and symbol visibility (only exports are visible).
6. **Cost and codes.** Which nodes change, and a code range (reserve E1610-E1639).

## Acceptance

- [ ] D122 written with Decision, Reason, Cost and Rejected, marked "Not built yet".
- [ ] Each of c-layout, c-export-functions and c-abi-export gets its brief amended in this node's outcome (the orchestrator applies it).

## Sizing

M: a design over the checker, emitter, runtime and driver, no shipped code.
