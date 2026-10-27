---
id: fmt
title: langc fmt
status: todo
priority: could
type: feature
outline: true
milestone: m7-tooling
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: split-compiler
    kind: hard
    why: needs the parser to keep comments, which touches files split-compiler settles
---

# langc fmt

## Goal

A formatter that defines the one valid layout, with `--check`.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
