---
id: selfhost-spike
title: Self-hosting gap check
status: todo
priority: must
type: feature
outline: true
milestone: m8-selfhost
target:
from:
size: M
spike: true
verify:
exclusive: []
fulfilled_by:
gate: true
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: modules
    kind: hard
    why: the compiler port is a multi-module package
  - on: maps
    kind: hard
    why: the port needs maps for symbol tables
  - on: strings-std
    kind: hard
    why: the lexer needs string positions
  - on: io-std
    kind: hard
    why: the driver reads files and runs the C compiler
  - on: box-recursive
    kind: hard
    why: the AST is recursive
  - on: patterns
    kind: hard
    why: the checker relies on nested patterns
  - on: closures
    kind: hard
    why: the port uses higher-order helpers
  - on: traits
    kind: hard
    why: symbol tables need Hashable keys
---

# Self-hosting gap check

## Goal

Port the lexer to Lang as a test, list the features still missing for the rest of the compiler, and estimate the port. The user reviews the list before the port starts.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
