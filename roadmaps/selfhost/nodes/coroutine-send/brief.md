---
id: coroutine-send
title: "Co.send for two-way generators"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: p3-review
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on: []
---

# Co.send for two-way generators

## Goal

D119: rename the two-argument `Co.resume(mutable co, value)` to `Co.send(mutable co, value)`, so no function is chosen by argument count. Update std/coroutine, the checker rows, SPEC, guide and samples.
