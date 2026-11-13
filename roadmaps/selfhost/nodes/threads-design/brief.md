---
id: threads-design
title: "Design: scoped threads, channels, Shared and Atomic"
status: active
priority: must
type: spike
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: S
spike: true
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/threads/**
  - roadmaps/selfhost/nodes/threads-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D111, diag E1190-E1199 and E1400-E1419]
provides: []
depends_on: []
---
# Design: scoped threads, channels, Shared and Atomic

## Goal

Decide how SPEC section 8 "Threads" is built on today's runtime, and write D111.

## Context

- SPEC section 8 "Threads" (around line 1381): `Thread.scope`, `Thread.start`, `Thread.join`, a `Scope` that cannot escape, channels (`Sender`, `Receiver`), `Shared<T>`, `Atomic<T>`. D80 (cycles through identity types), D90 (`Cell<T>`), D94 and D104 (shared writes and copies), D109 (one `_Thread_local` loop per thread).
- Reference counts are not atomic today, the size-class allocator (`narra_alloc`) keeps per-process free lists, and `Process.run` keeps one global child slot (see the threads outline notes).

## Questions

1. Which values may cross a thread boundary (a `Send`-like rule from types: no `Cell`, no `Async`, no `Generator`), and how the checker enforces it.
2. Atomic reference counts: everywhere, only for values reachable from `Shared`, or a deep copy at the boundary. Measure the cost on `bench/` and `emit selfhost`.
3. The allocator with threads (per-thread free lists, or a lock).
4. `Shared<T>` access rules (lock inside, `Shared.update(s, (mutable v) => ...)`), `Atomic<T>` types, channel buffering and closing.
5. Async with threads: one loop per thread, and passing work between loops.
6. Runtime fixes needed first (`Process.run` signals, test runner).
7. Split into build children (for example threads-runtime, threads-api, channels-shared) and refine their briefs; rename `nodes/threads/brief.md` as the first child or set it dropped with the children listed.

- Reserved codes: E1190-E1199 and E1400-E1419 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/threads/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D111: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D111 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D111 answers every question and is approved by the maintainer.
- [ ] The build brief follow D111, or the outcome says why the feature is dropped and which SPEC lines change.
