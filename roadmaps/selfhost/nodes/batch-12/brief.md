---
id: batch-12
title: "Batch 12: async, plus fixes from the batch-11 reviews"
status: active
priority: should
type: story
story: 
early: false
outline: false
milestone: m5-concurrency
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/batch-12/**
shared:
reserves: []
provides: []
depends_on:
  - on: batch-11
    kind: hard
    why: async builds on the batch-11 compiler and seed; the fixes target bugs found in its reviews
---
# Batch 12: async, plus fixes from the batch-11 reviews

## Goal

Twelfth one-shot batch, scope set by the maintainer on 2026-11-13: build async as D93 and D95 describe, and fix the bugs and nits the batch-11 reviews found. Children land on `feat/batch-12` as direct commits after a worker check and a fresh review. At the end: the seed updates, the full check and the Linux check, one fresh review of the whole branch, a bench row, and one pull request.

Children:

- async-design (S spike, gate D109): the runtime shape of `Async<T>`, tasks and the loop, the error codes, and SPEC section 8 rewritten to D93 and D95.
- async-core (L): `async function`, `async` arrows, `async { ... }`, `await`, the `Async<T>` type, the unused-handle error, an async entry function, and a ready-queue loop with no I/O.
- event-loop (M): kqueue and epoll, timers, `Async.sleep`, `Async.idle`, exit and cancel of tasks still running.
- async-api (M): `Async.then`, `detach`, `peek`, `cancel`, `all`, `race`.
- std-async-io (L, D109: counted `Handle` for sockets): `std/async/net`, `std/async/file` and `std/async/time`.
- poly-recursion-hang (S fix), arrow-arg-order (S fix), bound-trait-shared-results (S fix), shape-names-private (S fix).
- review-nits (S chore).

The async outline node is dropped, replaced by the async children.

## Order and shared files

- Machine: 8 GB, four building workers at a time. The spike and the fixes run beside async-design's gate.
- The fixes and review-nits do not wait for the gate. async-core waits for it; event-loop and async-api start from async-core (both change the runtime loop and `std/async`). The second to land rebases and keeps both sides. std-async-io starts from event-loop.
- The async runtime is C text in `selfhost/runtime/std_async.c`, `std_async_loop.c` and `std_async_io.c` (D109), each generated into its own module and picked in `runtimetext.nra`. `diagnostics.nra` is a table with one row per code, so each node appends its rows inside its own reserved codes and a rebase keeps both sides.
- Runtime text lives in `selfhost/src/runtime*.nra` and `selfhost/runtime/`; editing it needs two seed updates at the end of the batch, not per node.
- Every new syntax adds cases to the checker, emitter, fmt and astdump switches and rows to `diagnostics.nra`. Each feature's logic goes in its own module; shared files carry only dispatch.
- Codes (D109): async-core E1100-E1129, event-loop E1130-E1149, async-api E1150-E1169, std-async-io E1170-E1189, E1190-E1199 left for threads. Decision D109: async-design.
- Each worker uses a scratch folder named after its node under /private/tmp/claude-501/. Check a rebase's exit code before testing.

## Seed

`selfhost/src` and `std/src` may use only features the committed seed compiles, so no child uses async in the compiler or std sources it compiles with the seed. `std/async/*` modules are new std sources: they are compiled by the stage-1 compiler, and the seed update at the end makes them compile with the seed too.

## Acceptance

- [x] Every child is done, or moved out of the batch with a reason in this brief.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass on the story branch.
- [x] The seed is regenerated and `sh selfhost/bootstrap.sh` reports it current.
- [x] A debug-built compiler runs `emit selfhost` with `NARRA_CHECK_COPIES=1` and exits 0.
- [x] `compiler/README.md` "Not supported yet" no longer lists async functions or `async` arrows.
- [ ] A fresh review of the whole story branch, before the pull request.
- [ ] `bench/BASELINE.md` has an end-of-batch row; `narrac emit selfhost` instructions do not rise by more than the noise.
