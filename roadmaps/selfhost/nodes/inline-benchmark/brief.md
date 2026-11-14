---
id: inline-benchmark
title: "Spike: is @inline worth keeping"
status: todo
priority: should
type: spike
story:
early: false
outline: true
milestone: p3-review
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
gate: false
added: 2026-11-14
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on: []
---

# Spike: is @inline worth keeping

## Goal

D119: measure `@inline` on a quiet machine (bench programs and `emit selfhost`, with and without the attribute on std's Options.map, andThen, orElse and the compiler's hot helpers). If the gain is inside the noise, write a follow-up node that removes it.
