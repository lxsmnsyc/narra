---
id: node-addon
title: Node-API bindings for publishing Narra libraries to npm
status: todo
priority: later
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
  - on: c-abi-export
    kind: hard
    why: a `.node` addon is a shared library that exports `napi_register_module_v1`
---

# Node-API bindings for publishing Narra libraries to npm

## Goal

A Narra library (like napi-rs for Rust) that builds a Node-API addon from exported Narra functions: bindings for the Node-API C headers, conversion of strings, numbers, arrays and buffers, and errors thrown as JS exceptions. An example package publishes one npm package per platform (`os`, `cpu`, `libc` fields) plus a `wasm32-wasi` fallback, with a main package that lists them in `optionalDependencies` and loads the matching one.

## Notes

Parked. Its first user would be a TypeScript transformer written in Narra and called in-process from bundlers. Decide when refined whether this is a std module, a separate package, or only a guide with an example.
