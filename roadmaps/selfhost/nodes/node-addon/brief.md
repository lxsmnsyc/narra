---
id: node-addon
title: "Node-API glue and the .node output"
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
added: 2026-11-15
spent:
owns:
  - roadmaps/selfhost/nodes/node-addon/**
  - selfhost/src/nodeaddon.nra
  - testdata/node_addon/**
  - testdata/errors/node_addon*
shared:
  - selfhost/src/driver.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/*.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1640-E1649]
provides: []
depends_on:
  - on: c-abi-export
    kind: hard
    why: shared library output
  - on: c-export-functions
    kind: hard
    why: exports and the boundary wrapper
  - on: c-abi-design
    kind: hard
    why: D122 conversions, panics and threads
---
# Node-API glue and the .node output

## Goal

`narrac build --emit node-addon` writes `<name>.node` for the host or a `--target` row, as D123 item 3 says, with one generated Node-API wrapper per `@export(C)` function, `napi_register_module_v1`, and `<name>.d.ts`. Every export is synchronous. Types are those of D122 item 7: `bigint` for `int`, `i64` and `u64`, `Uint8Array` for `bytes`, and E1640 for `Ptr<T>`. The `.d.ts` follows it. The glue calls `narra_try_name` as a `static` function in the same C file. It does not take the lock itself: the entry does.

## Context

- T3 proof (`debate/js-interop/evidence/T3-node-addon.md`): about 30 lines of hand-written glue around emitted C built a 51 KB `.node`, and a call cost 0.18 µs.
- Emitted names are hashed statics today, so the T3 glue had to `#include` the C file. c-export-functions gives exports real symbols.
- This node replaces the earlier outline of the same name, dropped on 2026-11-14.

## Acceptance

- [ ] A sample library with `string`, `bytes`, number and `bool` exports loads with `require` in a test. The suite skips it with a message when `node` is missing. Each call returns the expected value, and a panic and an `Err` arrive as thrown `Error`s.
- [ ] A test loads the `.node` in two `worker_threads` and calls an export in both. A panic gives an `Error` with `code` `ERR_NARRA_PANIC`, and a later call works.
- [ ] The `.node` file exports only `napi_register_module_v1` (`nm`).
- [ ] The `.d.ts` file passes `tsc --noEmit` if TypeScript is on the machine. Otherwise the outcome says so.
- [ ] The leak check is clean after a million calls in one process.

## Sizing

M: a glue writer in the compiler, a driver output, and a Node test harness.
