---
id: remove-lazy
title: "Remove lazy values"
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

# Remove lazy values

## Goal

D119: remove `const x = lazy expr`, `std/lazy` and `Lazies.force`; retire E1032-E1035 (never reuse). Remove the lazyvalues and stdlazy modules, their emitter and effect cases, samples and SPEC text. No deprecation path.
