---
id: website-repl
title: REPL on the Narra website
status: todo
priority: could
type: feature
story:
early: false
outline: true
milestone: p6-docs
target:
ship: auto
from:
size: L
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
depends_on:
  - on: js-interop-design
    kind: hard
    why: D123 picks how the REPL compiles and runs Narra and writes its children
  - on: website
    kind: hard
    why: the REPL is a page of the site
---

# REPL on the Narra website

## Goal

A page on the website where a reader types Narra, sees diagnostics, and runs the program, so guide samples can open in it. Requested by the maintainer on 2026-11-14.

## Notes

Outline until js-interop-design writes D123; refine the children from it.
