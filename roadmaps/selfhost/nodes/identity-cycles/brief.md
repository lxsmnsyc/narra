---
id: identity-cycles
title: Reference cycles through identity types
status: active
priority: should
type: spike
story: batch-1
early: false
outline: false
milestone: m11-closures
target:
ship: pr
from:
size: S
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-01
spent:
owns:
  - roadmaps/selfhost/nodes/identity-cycles/**
shared:
  - SPEC.md
  - DECISIONS.md
reserves: [decision D80]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: the decision is checked against the self-hosted compiler's runtime and leak check (done)
---

# Reference cycles through identity types

## Goal

Fix how Narra treats reference cycles through identity types (`Shared<T>`, `Sender<T>`, `Receiver<T>`, `Task<T>`, and the planned `Cell<T>`), write it as D80, and correct SPEC section 6 "Cycles" and section 8 "Shared state". No identity type exists in the compilers yet (threads is parked, cell-capture is an outline), so this node changes only documents. cell-capture and threads build on the answer.

## Context

- SPEC section 6 "Cycles" says immutable values cannot form cycles, which still holds: an update copies when the count is above 1, and closure-borrow keeps `let` captures on the stack. Identity types break the claim. `Shared.update(cell, (mutable value) => { value = Some(cell) })` stores a handle in its own payload; a `Sender` sent into its own channel keeps the channel alive; a `Cell` holding a closure that captures the same `Cell` is the usual case once escaping closures share state.
- The leak check: `NARRA_CHECK_LEAKS=1` reports live allocations at exit (exit code 102), and `narra test` fails a leaking test in debug builds (SPEC section 13, `compiler/README.md`).

## Plan (D80)

1. No cycle collector and no type rule. A cycle through identity objects is a leak: memory safety holds, the memory is not freed, and resources in the cycle do not close.
2. Every cycle must pass through an identity object, since plain values cannot form one. So the debug leak check finds every such leak, and a test that leaks fails.
3. SPEC section 6 "Cycles" says this, with the two patterns that leak (a handle stored in its own payload, a closure in a `Cell` that captures the same `Cell`) and how to break them: clear the slot (`Cell.set(cell, None)`) before dropping the last outside handle, or keep graphs in `std/arena` with handles.
4. Section 8 "Shared state" and `Channel` say the same for their types. cell-capture and threads each add one leak sample whose expected result is the leak report, so the rule is tested where the types are built.
5. Leave room: a cycle collector limited to identity objects can be added later without changing any program, because it only frees memory that today's rule leaks.

## Acceptance

- [ ] DECISIONS.md D80 "Cycles through identity types leak and are caught by the leak check", with reasons, prior art (Swift and Rust reference counting, Nim ORC's collector), costs, and the alternatives below.
- [ ] SPEC section 6 "Cycles" and section 8 "Shared state" and "Channels" changed as in the plan; the bullet "A `let` variable or `mutable` parameter cannot be captured or stored" matches closure-borrow's wording if it has merged.
- [ ] The outcome lists what cell-capture and threads must test.

## Sizing

S. Documents only; the research is in this brief.

## Decision to confirm

Plan: documented leaks, caught by the debug leak check (option c of the outline, the orchestrator's earlier recommendation).

- Reason: zero cost for programs that use no identity types, and no pause or extra header for those that do. Frees and resource closes stay deterministic, the property SPEC section 6 promises.
- Reason: cycles can only pass through identity objects, which are created by hand and visible in the source, so a leak always has a named cause, and tests catch it.
- Rejected: a type rule that bans identity types inside identity payloads. It breaks request and reply channels (`Sender<#[Request, Sender<Reply>]>`) and stops a `Cell` from holding a closure, the main reason `Cell` exists.
- Rejected for now: a trial-deletion collector for identity objects (Nim ORC, CPython). Each identity decrement that leaves a count above zero would record a candidate, and objects shared across threads need a concurrent collector. It can be added later without breaking programs (point 5).
- Rejected: weak references (`Weak<T>`). They add a second handle kind for back edges that `std/arena` handles already cover.
