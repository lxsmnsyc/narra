---
id: c-seed
title: Bootstrap from a checked-in C seed and retire the Go compiler
status: todo
priority: later
type: refactor
story:
early: false
outline: true
milestone: m9-adoption
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-05
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: std-foundation
    kind: hard
    why: the stage 2 test path must exist before the Go tests can go
---

# Bootstrap from a checked-in C seed and retire the Go compiler

## Goal

Option (c) from the std-source gate: check in the C that stage 1 emits for the selfhost compiler as a seed, build the compiler from it with `cc` alone, and retire the Go compiler. Every language feature is then built once, in `selfhost/src`, instead of twice. The spike estimates the seed at about 102k lines and 4.6 MB of C.

## Notes

Outline, parked. Decide when refined: how the seed is regenerated and reviewed (it is generated, never edited), how the Go tests move to stage 2, where the compiler lives afterwards (this decides selfhost-layout's final directory), and whether to raise it to `should`, since every node built in both compilers costs about twice as much.
