---
id: identity-cycles
title: Reference cycles through identity types
status: todo
priority: should
type: spike
outline: true
milestone: m11-closures
target:
from:
size: M
spike: true
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: the decision is built and tested on the self-hosted compiler
---

# Reference cycles through identity types

## Goal

Decide how Narra handles reference cycles through identity types (`Shared<T>`, `Sender<T>`, `Receiver<T>`, `Task<T>`, and the planned `Cell<T>`). SPEC section 7 says immutable values cannot form cycles; identity types can: `Shared.update(cell, (mutable value) => { value = Some(cell) })`, or a `Sender` sent into its own channel. Options: a type rule that bans identity types inside identity payloads (breaks reply channels), a cycle collector limited to identity objects (trial deletion, as in Nim ORC), or documented leaks caught by the debug leak check, as in Rust. Write the decision record and correct SPEC section 7.

## Notes

Outline, gated: the user picks the option. The orchestrator recommended documented leaks first and a collector only if real programs need it. threads and cell-capture build on the answer.
