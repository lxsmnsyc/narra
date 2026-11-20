---
id: c-export-functions
title: "C-callable Narra functions: @export(C)"
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
reserves: [diag E1620-E1623]
provides: []
depends_on:
  - on: c-abi-design
    kind: hard
    why: D122 fixes the signatures, ownership, panic and init rules; D123 builds C libraries, Node addons and wasm glue on them
---
# C-callable Narra functions: @export(C)

## Goal

`@export(C, "name")` on a function gives it an unhashed C symbol with the C types of SPEC section 9, following D122: the clash check of c-interop (E1001 rules), two entries per export (the plain `name` and `narra_try_name`), as D122 item 3 says, with a panic stopped at the boundary as D122 says. No runtime state is set up on a call (D122 item 4), and ownership of strings and bytes as D122 says. Exported functions are reached roots for tree shaking.

## Context

The signature check covers only the types D122 allows without C layouts. Node c-layout is parked and is not on this path (D123 item 9).

This node owns the runtime work: `NARRA_SETJMP` and `NARRA_LONGJMP` macros (`_setjmp` and `_longjmp`) for every build, the panic location saved with the message, the UTF-8 check on `string` parameters, the library lock behind `NARRA_LIBRARY`, and `narra_<library>_alloc`, `_free` and `_shutdown`.

Codes: E1620 wrong place or arguments, E1621 type not allowed, E1622 type parameter, effect parameter, `async`, generator or `mutable` parameter, E1623 duplicate name or clash with an extern member. The name check reuses E1001.

SPEC section 9: delete the `@layout(C)` text and write the rules of D122.

## Acceptance

- [ ] A sample with a C source file (`[link]` or a test harness step) that calls an exported Narra function as a callback (for example `qsort` with an exported comparator; an executable with `@export` functions takes no lock) runs clean under AddressSanitizer and the leak check.
- [ ] A panic inside an exported function does what D122 says, tested in the exits suite.
- [ ] The export wrapper is the function the glue of node-addon and wasm-js-glue calls. A panic inside an export called through it is stopped at the boundary as D122 says, tested in the exits suite.
- [ ] Errors in E1620-E1623 for signatures D122 rejects.
- [ ] The checked entry returns 2 and the message for a panic, 1 and the message for an `Err`, and a bad UTF-8 argument gives status 1, in the exits suite.
- [ ] A million calls of an export with `string` and `bytes` leave the live count at 0 under `NARRA_CHECK_LEAKS` after `shutdown`.
- [ ] A bench line for the checked entry on macOS shows the guard under 15 ns.
- [ ] `sh scripts/test.sh` passes.

## Sizing

M: one attribute, a wrapper per export, the boundary rules.
