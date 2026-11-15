---
id: website-repl
title: REPL on the Narra website
status: todo
priority: later
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

Parked by the maintainer on 2026-11-14. D123 keeps the measurements (narrac in wasm, clang in the page); refine from it when the REPL is picked up.


D123 is written and parks the REPL: no children yet. If it returns, check, `fmt` and the emitted C run in the page with the wasm build of narrac (`scripts/wasm-build.sh`). Running a program loads clang built for wasm on first press (about 19.8 MB) or uses a server. That choice is not made. The orchestrator sets priority later.
