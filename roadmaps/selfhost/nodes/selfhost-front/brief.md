---
id: selfhost-front
title: Self-hosted lexer and parser
status: todo
priority: must
type: feature
outline: true
milestone: m8-selfhost
target:
from:
size: L
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
  - on: selfhost-spike
    kind: hard
    why: needs the gap list and any features it adds
---

# Self-hosted lexer and parser

## Goal

Lexer, AST and parser in Narra, producing the same tree as the Go parser on every test program.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.
