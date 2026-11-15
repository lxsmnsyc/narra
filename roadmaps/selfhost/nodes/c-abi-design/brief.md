---
id: c-abi-design
title: "Spike: ownership and runtime rules at the C library boundary"
status: todo
priority: could
type: spike
story: batch-16
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
    why: D123 builds C libraries, Node addons and wasm glue on this boundary (maintainer, 2026-11-15)
---
# Spike: ownership and runtime rules at the C library boundary

## Goal

Write decision D122, in the style of D109, for C interop step 2 (SPEC section 9, "Not built yet"): C-callable Narra functions and library output, the base for C libraries, Node addons and the wasm glue of D123. C layouts only if they are on that path. The maintainer reads it before c-export-functions, c-abi-export, node-addon or wasm-js-glue start (gate).

## Questions

1. **Ownership across the boundary.** Strings and bytes passed in: borrowed for the call, or copied. Returned: the library allocates, the host frees through one exported free function. Which reference-counted types may appear in an exported signature at all, and whether an opaque handle type (`Ptr<T>` to a Narra value, with exported retain and release) is the way to hand out structs that hold strings or arrays. Nothing may leak, and a host must never free memory it cannot free.
2. **Layouts.** SPEC names `@repr(C)` and `@layout(C)` in different places. Pick one spelling for structs and one for enums with fixed integer values, and the field types allowed (numbers, `bool`, `Ptr<T>`, nested C structs, fixed arrays).
3. **Exported functions.** `@export(C, "name")` rules: allowed signatures, no type parameters, effects (does a C caller need `unsafe`?), name clash check against the name table (E1001 rules from c-interop), and whether an exported function may also be called from Narra.
4. **Runtime state in a library.** Initialization without `main` (lazy on first call, or an exported init), per-thread heaps (D111) when an unknown host thread calls in, the async loop, and what a panic does (abort the process, or return an error code through an out parameter). A panic must never unwind into the host.
5. **Output.** `--emit shared-lib` and `static-lib` names per target (`.dylib`, `.so`, `.dll` with an import lib, `.a`, wasm reactor module), `--emit node-addon` (`.node`), the generated header, and symbol visibility (only exports are visible).
6. **Cost and codes.** Which nodes change, and a code range (E1610-E1639 stay reserved; node-addon has E1640-E1649 and wasm-js-glue E1650-E1659).
7. **Node-API.** Which types the generated glue converts in the first build (numbers, `bool`, `string`, `bytes`) and how. What a panic and an `Err` become: a thrown JS `Error` with the message, with the panic stopped at the boundary by the runtime's `setjmp` guard, not ending the process. Whether a later build adds arrays, `Option` (as `undefined`) and plain structs as JS objects.
8. **Node threads.** An addon may load in several `worker_threads`. Say whether runtime state is per Node environment (Node-API instance data) or per thread (D111 heaps), and what an export called on a worker does.
9. **Wasm reactor.** The exports built for `Wasi-Wasm32` as a reactor module with no `main`, and the memory calls the JS glue of wasm-js-glue uses to pass strings and bytes.

## Acceptance

- [ ] D122 written with Decision, Reason, Cost and Rejected, marked "Not built yet".
- [ ] Each of c-export-functions, c-abi-export, node-addon and wasm-js-glue gets its brief amended in this node's outcome (the orchestrator applies it). c-layout stays parked unless D122 puts C layouts on the path of a library.

## Sizing

M: a design over the checker, emitter, runtime and driver, no shipped code.
