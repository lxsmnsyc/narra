---
id: selfhost-check
title: Self-hosted checker
status: dropped
priority: must
type: feature
outline: true
milestone: m8-selfhost
target:
from:
size: XL
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
  - on: selfhost-front
    kind: hard
    why: checks the AST the front end produces
---

# Self-hosted checker

## Goal

Type checker in Narra, including generics, traits and patterns. Split before starting.

## Notes

Outline. Break into full briefs when `plan.py refine` lists it.

## Dropped

2026-10-30: split into selfhost-check-decls and selfhost-check-bodies.
