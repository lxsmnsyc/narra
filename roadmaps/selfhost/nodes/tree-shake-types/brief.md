---
id: tree-shake-types
title: "Tree shaking: types, helpers and runtime sections"
status: todo
priority: should
type: feature
story: batch-7
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/tree-shake-types/**
shared:
  - selfhost/src/emit.nra
  - selfhost/src/emitter.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitmaps.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/emitclosures.nra
  - selfhost/src/emitstate.nra
reserves: []
provides: []
depends_on:
  - on: tree-shake
    kind: hard
    why: uses its reachable set
---

# Tree shaking: types, helpers and runtime sections

## Goal

After tree-shake, the emitter still writes every type the checker collected and every helper for it: typedefs and structs, array and map instances, reference counting, eq/hash/cmp/show, copy-check credit functions, the comment table, string literals, and the runtime section of every imported std module (`runtimeFor` in `emit.nra` keys on `checker.used`, the imports). Emit only those that reachable code uses.

## Context

- Collect the types reachable from the emitted functions' signatures, locals and expressions, closed over fields, payloads and element types. Every per-type section filters on that set.
- A std runtime section stays when a reachable function calls a builtin from it.
- The `Option<int>`, `#[int, int]` instances that `fib` gained in batch-6 (from the ext functions) are the first check.

## Acceptance

- [ ] `narrac emit bench/fib/fib.nra` is no longer than at the start of batch-6 (1012 lines with the batch-5 compiler), apart from runtime text added since.
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point; the outcome records lines of `narrac emit selfhost` before and after.
- [ ] Fast check passes: `sh scripts/test.sh`
