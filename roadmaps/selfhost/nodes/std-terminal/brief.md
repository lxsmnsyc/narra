---
id: std-terminal
title: "std/terminal"
status: todo
priority: should
type: feature
story: batch-14
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
added: 2026-11-13
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: the design gate fixes the API this node builds

---

# std/terminal

## Goal

Colors and styles, TTY detection and terminal width, for diagnostics and CLI output. Respects NO_COLOR.

## Notes

Outline, added 2026-11-13 when the maintainer set the phase order (syntax, std, feature review, packages, distribution, docs).
