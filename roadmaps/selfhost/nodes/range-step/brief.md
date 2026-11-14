---
id: range-step
title: "Ranges.stepped(start, end, step)"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: p2-std
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

# Ranges.stepped(start, end, step)

## Goal

D119: a stepped range in std/range for counting down and strides. A constant step of 0 is a compile error; a run-time step of 0 gives an empty range. Works with for-of and std/iterator.
