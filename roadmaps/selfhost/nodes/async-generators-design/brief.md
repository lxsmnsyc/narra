---
id: async-generators-design
title: "Design: async generators and for await"
status: done
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
spent: 161270
owns:
  - roadmaps/selfhost/nodes/async-generators/**
  - roadmaps/selfhost/nodes/async-generators-design/**
shared:
  - DECISIONS.md
  - SPEC.md
  - selfhost/src/diagnostics.nra
reserves: [decision D110]
provides: []
depends_on: []
---
# Design: async generators and for await

## Goal

Decide how `async function*`, `AsyncGenerator<T>` and `for await (const chunk of stream)` work, as D95 asks ("needed for HTTP and streaming"), and write D110.

## Context

- D95 "Async exit" bullet, D109 (the async runtime: task frames reuse the generator state machine, cooperative cancel down the await chain), D31 (iterators are generators), D88 (generator blocks parked).
- SPEC section 8 says `async function*` is "not present in the first version" (around lines 311 and 1531).
- Code: `selfhost/src/emitgenerators.nra`, `emitasync*.nra`, `selfhost/runtime/std_async*.c`, `std/src/async.nra`.

## Questions

1. Type and protocol: `AsyncGenerator<T>` as a builtin like `Generator<T>`; what `next` returns (an `Async<Option<T>>` or a suspend inside `for await`).
2. Where `for await` is allowed (async bodies only), and whether it works over `Generator<T>` too.
3. `yield` and `await` in one body; `yield*` over another async generator.
4. Cancel mid-iteration (D109 item 8), early `break`, and what is released.
5. Backpressure: the producer runs only when the consumer asks.
6. Effects, `ensures` on each yielded value (D95), the D85 rule in `for await`.
7. Whether std-async-io gets stream functions (`Net.chunks`, `File.lines`) in this node or later; the time operators from the inbox (`debounce`, `throttle`, `merge`) stay out unless trivial.

- Reserved codes: E1112-E1129 (add the range to the `diagnostics.nra` header comment and split it between the build briefs).
- Build brief to refine: `nodes/async-generators/brief.md`.

## Steps

- [ ] Read the sources above and the code the decision changes.
- [ ] Answer the questions in D110: decision, reason, cost, rejected options. Mark it "Proposed, waiting for the maintainer".
- [ ] Write the SPEC text the decision changes, marked as not built yet where the build node has not landed.
- [ ] Refine the build brief listed below from the template: owns, shared, typed `depends_on` (this spike, hard), codes from the reserved range, acceptance with samples and error tests, README changes, `verify: sh scripts/test.sh`. Split any brief larger than L.
- [ ] Hand D110 to the maintainer. Build nodes start only after approval.

## Acceptance

- [ ] D110 answers every question and is approved by the maintainer.
- [ ] The build brief follow D110, or the outcome says why the feature is dropped and which SPEC lines change.
