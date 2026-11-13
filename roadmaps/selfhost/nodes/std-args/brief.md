---
id: std-args
title: "std/args"
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
size: M
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

# std/args

## Goal

Command-line argument parsing for narrac and user tools: flags, options with values, subcommands, positional arguments, help text and errors as values.

## Notes

Outline, added 2026-11-13 when the maintainer set the phase order (syntax, std, feature review, packages, distribution, docs).
