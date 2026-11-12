---
id: distinct-traits
title: "implements and derives on distinct types"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-12
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on: []
---

# implements and derives on distinct types

## Goal

SPEC allows a distinct type to list traits with `implements` or `derives`; the compiler rejects it (README, "Differences from the spec"). Also decide whether `==` on a distinct type through its base type stays (the README lists it as a difference) or follows SPEC.

## Notes

- From the language-first triage on 2026-11-12: `compiler/README.md` lists it under "Not supported yet" or "Differences from the spec". Check it against the current compiler before refining.
- Outline. Break into a full brief when `plan.py refine` lists it.
