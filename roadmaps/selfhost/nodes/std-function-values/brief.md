---
id: std-function-values
title: "Standard functions as values"
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
size: M
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

# Standard functions as values

## Goal

Pass a std function such as `Int.toString` where a function value is expected, as SPEC allows (README, "Not supported yet"). Std functions are builtins with their own check and emit paths, so this needs a wrapper the emitter can take the address of, with the right effects in its function type.

## Notes

- From the language-first triage on 2026-11-12: `compiler/README.md` lists it under "Not supported yet" or "Differences from the spec". Check it against the current compiler before refining.
- Outline. Break into a full brief when `plan.py refine` lists it.
