---
id: generic-field-defaults
title: "Defaults for generic struct fields"
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

# Defaults for generic struct fields

## Goal

A struct field whose type uses a type parameter cannot have a default (README, "Not supported yet"). Check the default once per instance, or require it to type-check for every `T`, as SPEC says.

## Notes

- From the language-first triage on 2026-11-12: `compiler/README.md` lists it under "Not supported yet" or "Differences from the spec". Check it against the current compiler before refining.
- Outline. Break into a full brief when `plan.py refine` lists it.
