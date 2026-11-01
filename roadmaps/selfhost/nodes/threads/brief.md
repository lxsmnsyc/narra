---
id: threads
title: Scoped threads, channels and shared state
status: todo
priority: later
type: feature
outline: true
milestone: m5-concurrency
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: closures
    kind: hard
    why: thread functions are escaping closures
---

# Scoped threads, channels and shared state

## Goal

`Thread.scope`, `Thread.start`, channels, `Shared<T>`, `Atomic<T>`, and atomic reference counts for values shared across threads.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.

Cycles (added 2026-11-01): identity types can form reference cycles that never free. Examples: `const cell = Shared.create(None)` then `Shared.update(cell, (mutable value) => { value = Some(cell) })`, or sending a `Sender` into its own channel. Pick one before refining: a type rule that a `Shared`, channel or task payload cannot hold identity types (breaks reply channels), a cycle collector limited to identity types (trial deletion, as in Nim ORC), or documented leaks like Rust caught by the debug leak check. Update SPEC section 7 to match.
