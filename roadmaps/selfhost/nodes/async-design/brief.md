---
id: async-design
title: "Design: the async runtime, codes and SPEC text"
status: todo
priority: should
type: spike
story: batch-12
early: false
outline: false
milestone: 
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
  - roadmaps/selfhost/nodes/async-core/**
  - roadmaps/selfhost/nodes/event-loop/**
  - roadmaps/selfhost/nodes/async-api/**
  - roadmaps/selfhost/nodes/std-async-io/**
  - roadmaps/selfhost/nodes/async-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D109]
provides: []
depends_on: []
---
# Design: the async runtime, codes and SPEC text

## Goal

D93 and D95 settle what async means to a program. They do not settle how the compiler and runtime build it. Write D109 with those answers, rewrite the "Async functions" part of SPEC section 8 to match D93 and D95 (the current text still describes D17: `Async.run`, `Async.spawn`, `Task<T>`, `Loop`), and refine the briefs of async-core, event-loop, async-api and std-async-io from it. The maintainer approves D109 and the SPEC text before async-core starts.

## Context

- D93 (DECISIONS.md around line 955) and D95's "Async exit" bullet (around line 1021) are the design. D88 parks generator blocks; async blocks are reopened by D93.
- Generators compile to a frame plus resume, clone and drop functions (`selfhost/src/emitgenerators.nra`, header comment). Coroutines are in `selfhost/src/stdcoroutine.nra` and `std/src/coroutine.nra`. The runtime is C text in `selfhost/src/runtime*.nra` and `selfhost/runtime/`.
- `Async<T>` is an identity type (D80), like `Cell<T>` (D90, `selfhost/src/stdcell.nra`).
- `async` is not a lexer keyword; the parser rejects a nested `async function` with E0930 (`selfhost/src/parser.nra` around line 2675) and parses no other async form. `compiler/README.md` lists async under "Not supported yet" (around lines 221-230).

## Questions D109 must answer

Each has a recommended answer. Check it against the code and change it if the evidence says so.

1. **Body transform.** Recommended: an async body reuses the generator state machine; `await` is a suspend point that registers the task on the awaited handle and returns to the loop. Say which parts of `emitgenerators.nra` are shared and which are new.
2. **Handle layout.** Recommended: one counted heap object per handle with state (pending, done, cancelled), the value slot, the `then` callbacks and the waiting tasks. Releasing the last handle of a finished task frees the value.
3. **Where the async runtime lives.** One source per part (`selfhost/runtime/async.c`, `async_loop.c`, `async_io.c`), each generated into its own `runtime*.nra` module, so the nodes do not share one file. Say whether `Async<T>` is a builtin type like `Type.TGenerator` or a std type special-cased like `Cell`.
4. **Where `await` is allowed.** Only in async bodies (functions, arrows, blocks). State the error codes and what `await` in a generator body or a non-async arrow reports.
5. **Unused handle.** Which uses count (assigned, passed, returned, awaited, detached) and which code reports the rest.
6. **Entry function.** How an `async` entry is called, how its `Async<int>` gives the exit code, and that running tasks are cancelled at exit (D95).
7. **Loop and I/O.** One loop per thread in the runtime, kqueue on macOS and epoll on Linux, a timer heap, and what happens on wasm32-wasi (stub or not supported). Where blocking std calls inside async code stand (allowed, they block the loop).
8. **Cancel.** How `Err(Cancelled)` reaches the I/O a task waits on next, and the error type std-async-io returns.
9. **Effects and copies.** Effects charged at creation (D93). How captured values move into the task frame under D94 and D104.
10. **Debug checks.** How the leak check and AddressSanitizer see tasks still pending at exit.
11. **Codes.** Add the range E1100-E1199 (concurrency) to the `diagnostics.nra` header (header comment only) and split codes between the four async nodes in their briefs.

## Steps

- [ ] Read D93, D95, `emitgenerators.nra` and the runtime parts D109 changes.
- [ ] Answer the eleven questions in D109: decision, reason, cost, rejected options.
- [ ] Rewrite the "Async functions" part of SPEC section 8. Keep the examples TypeScript-like and runnable once the nodes land.
- [ ] Refine the four async briefs: owns, shared, codes, acceptance with samples and error tests. Split any that is larger than L.
- [ ] Hand D109 and the SPEC text to the maintainer.

## Acceptance

- [ ] D109 answers the eleven questions and is approved by the maintainer.
- [ ] SPEC describes D93 and D95, with no `Async.run`, `Async.spawn`, `Task<T>` or `Loop` left anywhere in the file (today also around lines 234, 308, 310, 1297).
- [ ] The briefs of async-core, event-loop, async-api and std-async-io follow D109.
