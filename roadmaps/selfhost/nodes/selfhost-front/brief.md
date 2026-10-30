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

- Design (user decision 2026-10-30): the parser gives every expression, pattern and local an integer id, so the checker can keep results in side tables (selfhost-spike outcome, gap 3). The id scheme is part of this node.
- Error flow: no recover in Narra; keep a `failed` field in parser state (gap 7).
Outline. Break into full briefs when `plan.py refine` lists it.
