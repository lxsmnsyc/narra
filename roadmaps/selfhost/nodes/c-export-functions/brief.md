---
id: c-export-functions
title: "C-callable Narra functions: @export(C)"
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
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/c-export-functions/**
  - selfhost/src/cexports.nra
  - testdata/c_export*
  - testdata/errors/c_export*
shared:
  - selfhost/src/parser.nra
  - selfhost/src/externs.nra
  - selfhost/src/emitexterns.nra
  - selfhost/src/diagnostics.nra
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimecore.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1620-E1629]
provides: []
depends_on:
  - on: c-abi-design
    kind: hard
    why: D122 fixes the signatures, ownership, panic and init rules; D123 builds C libraries, Node addons and wasm glue on them
---
# C-callable Narra functions: @export(C)

## Goal

`@export(C, "name")` on a function gives it an unhashed C symbol with the C types of SPEC section 9, following D122: the clash check of c-interop (E1001 rules), a wrapper that sets up runtime state on first call and stops a panic at the boundary as D122 says, and ownership of strings and bytes as D122 says. Exported functions are reached roots for tree shaking.

## Context

The signature check covers only the types D122 allows without C layouts. Node c-layout is parked and is not on this path (D123 item 9).

## Acceptance

- [ ] A sample with a C source file (`[link]` or a test harness step) that calls an exported Narra function as a callback (for example `qsort` with an exported comparator) runs clean under AddressSanitizer and the leak check.
- [ ] A panic inside an exported function does what D122 says, tested in the exits suite.
- [ ] The export wrapper is the function the glue of node-addon and wasm-js-glue calls. A panic inside an export called through it is stopped at the boundary as D122 says, tested in the exits suite.
- [ ] Errors in E1620-E1629 for signatures D122 rejects.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: one attribute, a wrapper per export, the boundary rules.
