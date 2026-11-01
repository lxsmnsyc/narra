---
id: cell-capture
title: Cell<T> for mutable state in escaping closures
status: todo
priority: should
type: feature
outline: true
milestone: m11-closures
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: identity-cycles
    kind: hard
    why: a Cell can hold a closure that captures the same Cell, so the cycle decision comes first
  - on: closure-borrow
    kind: hard
    why: shares the capture rules and checker code that closure-borrow adds
---

# Cell<T> for mutable state in escaping closures

## Goal

`Cell<T>`, an identity type for mutable state that escaping closures share: `const count = Cell.create(0)` captured as a constant, read and changed with `Cell.get` and `Cell.set` or an update callback. The heap cell is created by hand, so the cost stays visible (D46). A `Cell` cannot be sent to another thread (use `Shared`). Implicit capture of a `let` by escaping closures stays an error.

## Notes

Outline. Decide when refining whether `Cell` lives in `std/cell` and how it relates to `Shared` (same API without the lock).
