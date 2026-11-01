---
id: closure-borrow
title: Non-escaping closures capture let and mutable by reference
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
  - on: selfhost-bootstrap
    kind: hard
    why: language changes after bootstrap go into the Narra compiler only
---

# Non-escaping closures capture let and mutable by reference

## Goal

A closure passed directly to a non-escaping function parameter may read and assign `let` variables and `mutable` parameters of the enclosing function, by reference: `let total = 0; Arrays.forEach(items, (item) => { total = total + item })`. Sound with today's rules: a non-escaping environment lives on the caller's stack (D46), so the pointer cannot dangle.

## Notes

Outline. Rules to write into the decision record and SPEC sections 2 and 6: a variable the closure writes is borrowed for that call, so the same call cannot also pass it as `mutable` or read it in another argument; a closure stored, returned, captured by an escaping closure or passed through a function value (D51) keeps today's error; effects of the closure flow to the call as now. Prior art: Swift lets non-escaping closures capture `inout` parameters. Add samples and compile-error tests, and use it in the selfhost compiler where loops worked around the ban.
