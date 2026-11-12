---
id: tuple-index
title: "Reading a tuple element by index"
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

# Reading a tuple element by index

## Goal

SPEC allows reading one element of a tuple by its index; the compiler does not (README, "Not supported yet"). Parser, checker (the index must be a literal in range), emitter, fmt and astdump, plus samples and error tests.

## Notes

- From the language-first triage on 2026-11-12: `compiler/README.md` lists it under "Not supported yet" or "Differences from the spec". Check it against the current compiler before refining.
- Outline. Break into a full brief when `plan.py refine` lists it.
