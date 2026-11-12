---
id: generic-implement
title: "implement blocks and constraints on generic types"
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
size: L
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

# implement blocks and constraints on generic types

## Goal

Today `implement` works only for non-generic structs and enums, and constraints are allowed only on the type parameters of functions (README, "Differences from the spec"). Allow `implement Trait for Box<T>` style blocks on generic types, and constraints on the type parameters of structs and enums, as SPEC describes. Likely split: constraints first, then generic `implement`.

## Notes

- From the language-first triage on 2026-11-12: `compiler/README.md` lists it under "Not supported yet" or "Differences from the spec". Check it against the current compiler before refining.
- Outline. Break into a full brief when `plan.py refine` lists it.
