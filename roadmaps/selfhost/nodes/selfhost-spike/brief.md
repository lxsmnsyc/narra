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

Port the lexer to Narra as a test, list the features still missing for the rest of the compiler, and estimate the port. The user reviews the list before the port starts.

## Notes

Facts from wave 3: C names are hashed (xxHash32 needs `*%` and rotates), fielded enums are heap payloads, `Box<T>` and std/fs, std/process, std/env exist. User-code recursion is limited to about 20000 frames; check that the lexer and parser port do not recurse per character or per token. Outline. Break into full briefs when `plan.py refine` lists it.
