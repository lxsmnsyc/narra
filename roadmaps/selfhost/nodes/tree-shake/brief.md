---
id: tree-shake
title: "Tree shaking: emit only reachable declarations"
status: done
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
spent: 327830
owns:
  - roadmaps/selfhost/nodes/tree-shake/**
shared:
  - selfhost/src/emit.nra
  - selfhost/src/emitter.nra
  - selfhost/src/emitdecls.nra
reserves: []
provides: []
depends_on:
  - on: copy-check-macros
    kind: hard
    why: both edit emitdecls.nra (the credit functions tree-shake filters)
  - on: build-conditions
    kind: hard
    why: the reachable set must skip the branch a build condition drops, so it reads the fold build-conditions adds
---

# Tree shaking: emit only reachable declarations

## Goal

Whole-program tree shaking in the emitter, for user and std modules alike. Today every function, constant and function body the checker lists is emitted: `bench/fib` carries `compare`, `divmod` and `floorDiv` from `std/src/ext__int.nra` without calling them, and the share grows as std moves to Narra source. Compute the declarations reachable from the program's roots and emit only those.

## Context

- Roots: the entry function (`findEntry`, `emitter.nra`), or every test block in a test build; top-level constants read by reachable code.
- Edges, recorded once per body (the checker's side tables already resolve every reference): direct calls, function values, arrow functions and their environments, generator and coroutine bodies, trait implementations a reachable call or derived function can dispatch to (eq, hash, cmp, show and user traits; take every implementation of a trait for a type that is reachable, rather than tracking dispatch exactly), and generic instances.
- Everything is still type-checked; only emission changes. Diagnostics and the language server are untouched.
- Builds stay deterministic (SPEC section 12): emit in the current order, filtered; never in set order.
- A missing root or edge shows up as a C compile or link error, so the full suite and the bootstrap fixed point are the main tests.
- tree-shake-types removes the types and helpers that only unreachable code needed; this node does functions, constants and bodies.

## Acceptance

- [ ] `narrac emit bench/fib/fib.nra` has no `ext__int` function bodies, and a sample with an unused user function, an unused generic instance and an unused trait impl emits none of them.
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point; the outcome records lines of `narrac emit selfhost` and `emit` instructions before and after.
- [ ] Fast check passes: `sh scripts/test.sh`
