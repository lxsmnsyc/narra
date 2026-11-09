---
id: cell-capture
title: "Cell<T> for mutable state in escaping closures"
status: todo
priority: should
type: feature
story: batch-2
early: false
outline: false
milestone: m11-closures
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-01
spent: 203844
owns:
  - std/src/cell*
  - testdata/cell*
  - testdata/errors/cell*
  - roadmaps/selfhost/nodes/cell-capture/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [decision D90, code E0965, code E0966]
provides: []
depends_on:
  - on: identity-cycles
    kind: hard
    why: a Cell can hold a closure that captures the same Cell, so the cycle decision (D80) comes first
  - on: closure-borrow
    kind: hard
    why: shares the capture rules and checker code that closure-borrow adds
---

# Cell<T> for mutable state in escaping closures

## Goal

`Cell<T>`, an identity type for mutable state that escaping closures share. `const count = Cell.create(0)`, captured as a constant, read and changed with `Cell.get`, `Cell.set` and `Cell.update(count, (n) => n + 1)`. The heap cell is created by hand, so the cost stays visible (D46). Implicit capture of a `let` by escaping closures stays an error, and its message now suggests `Cell`.

SPEC.md already names `Cell<T>` as planned in sections 6 and 8 (cycles through identity types, D80). Read those lines and D80 first.

Decide, and write as D90 (the orchestrator makes the gate decision):

- Module: `std/cell` as a source module over a `std/internal` primitive, or a builtin like `Shared`. Prefer source if std-foundation's rules allow it.
- API names, matching `Shared` without the lock where they overlap.
- `Cell` cannot be sent to another thread (use `Shared`); how the checker says so today, given threads are parked.
- Equality: identity types have no structural equality; confirm.

## Acceptance

- [ ] D90 in DECISIONS.md; SPEC.md sections 6 and 8 describe `Cell` as built.
- [ ] Samples `testdata/cell*.nra` with `.out`, clean under AddressSanitizer and the leak check: a counter shared by two escaping closures, `Cell.update`, a `Cell<string>` and a `Cell<Option<...>>`.
- [ ] A sample with a cycle through a `Cell` (a closure stored in the cell that captures it) fails the leak check as D80 says, and passes after `Cell.set(cell, None)`. This may be a test in the leak-check suite rather than a sample.
- [ ] Error tests `testdata/errors/cell*`: escaping capture of a `let` suggests `Cell`; misuse of the API.
- [ ] `sh scripts/test.sh` passes.
