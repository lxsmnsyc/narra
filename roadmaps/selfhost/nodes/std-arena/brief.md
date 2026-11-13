---
id: std-arena
title: std/arena
status: todo
priority: should
type: feature
story: batch-14
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - std/src/arena.nra
  - selfhost/src/stdarena.nra
  - testdata/arena*.nra
  - testdata/arena*.out
  - testdata/errors/arena_*
  - roadmaps/selfhost/nodes/std-arena/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/stdtext.nra
  - testdata/errors/unknown_module.err
  - testdata/declsdump/imports/err_unknown_module.dump
  - SPEC.md
  - compiler/README.md
  - docs/guide/
reserves: [diag E1310-E1319]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 2 fixes the types and functions
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests (std-source gate, 2026-11-05)
  - on: std-source
    kind: hard
    why: the module is Narra source
---

# std/arena

## Goal

Build `std/arena` (alias `Arenas`) as D118 item 2 says: `Arena<T>` and `Handle<T>` with generation-checked handles, the answer for graphs and cycles under reference counting (SPEC section 6, Cycles). Pure Narra source over `Array`.

## Context

- Source: D118 item 2; SPEC section 3 "Std modules for tools" and section 6 "Cycles".
- Evidence: a prototype with `struct Handle<T> derives Equatable, Hashable { index: int, generation: int }`, a phantom `T`, used as a map key, builds and passes `NARRA_CHECK_LEAKS=1` on the batch-14 seed. `get` needs `from arena` on its result (D94).

## Steps

- [ ] `export opaque struct Arena<T>` (slots and a free list) and `export opaque struct Handle<T> derives Equatable, Hashable, Show`.
- [ ] `create`, `insert`, `get` (`Option<T> from arena`), `set` and `update` (`bool`, false for a stale handle), `remove` (`Option<T>`), `contains`, `size`, `clear` (raises every generation) and `entries` (`Generator<#[Handle<T>, T]>`).
- [ ] Registry row, `stdlib.nra`, regenerated `stdtext.nra`, module-list goldens.
- [ ] SPEC: the built text in section 3 and the "Cycles" bullet without "not built yet"; the guide's standard library chapter.

## Acceptance

- [ ] `testdata/arena.nra`: insert, get, update, remove, reuse of a slot (the old handle gets `None` and `false`), `clear`, `entries` order, and a handle as a map key.
- [ ] `testdata/arena_graph.nra`: a graph with a cycle of handles is freed when the arena goes; clean under the leak check and AddressSanitizer.
- [ ] Error tests: `testdata/errors/arena_handle_fields.nra` (reading `handle.index` outside the module; opaque) and `testdata/errors/arena_wrong_type.nra` (a `Handle<int>` passed to an `Arena<string>`).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

S: about 200 lines of Narra and two samples.

## Notes for the worker

- If generic opaque structs or `derives Show` on a phantom type parameter do not work, record it in `log.md` and ask the orchestrator before working around it; do not change the checker in this node.
- A handle from another arena is not detected (D118 cost). Say so in the SPEC text.
- Codes E1310-E1319 are reserved; none is expected.
