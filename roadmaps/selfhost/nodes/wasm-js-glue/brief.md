---
id: wasm-js-glue
title: "JS wrappers for a wasm library, written by narrac"
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
  - roadmaps/selfhost/nodes/wasm-js-glue/**
  - selfhost/src/wasmglue.nra
  - testdata/wasm_js/**
  - testdata/errors/wasm_js*
shared:
  - selfhost/src/driver.nra
  - selfhost/src/diagnostics.nra
  - selfhost/tests/src/*.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1650-E1659]
provides: []
depends_on:
  - on: c-abi-export
    kind: hard
    why: reactor output
  - on: c-abi-design
    kind: hard
    why: D122 conversions
  - on: wasi-runtime
    kind: hard
    why: the wasi runtime the module imports
  - on: node-addon
    kind: behavior
    why: same sample and same `.d.ts`
---
# JS wrappers for a wasm library, written by narrac

## Goal

`narrac build --emit wasm-js --target Wasi-Wasm32` writes the reactor module and `<name>.js` plus `<name>.d.ts`, as D123 item 4 says. The JS module instantiates the reactor, holds the WASI preview1 calls the runtime imports, copies strings and bytes through the D122 allocator calls, and throws a JS `Error` for a panic or an `Err`. No outside npm package (no emnapi).

## Acceptance

- [ ] node-addon's sample gives the same results through `<name>.js` in Node as through the `.node` build, and in a headless browser if one is on the machine. Otherwise the outcome says so.
- [ ] The generated JS imports nothing outside itself, and `<name>.d.ts` matches node-addon's.
- [ ] A panic inside an export leaves the instance usable for the next call.

## Sizing

M: a JS writer in the compiler and a memory protocol shared with the D122 reactor.
