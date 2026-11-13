---
id: constructor-field-defaults
title: "A qualified enum constructor is not a constant default"
status: active
priority: must
type: fix
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/constructor-field-defaults/**
shared:
  - selfhost/src/exprs.nra
  - selfhost/src/effects.nra
reserves: []
provides: []
depends_on: []
---

# A qualified enum constructor is not a constant default

## Goal

`isConstantValue` in `selfhost/src/exprs.nra` (around line 3069) accepts an `Ident` that refers to a constructor but not a qualified one, so a struct field default `rest: Rest = Rest.None` is E0106, though enum constructors are always qualified. Found by std-args (batch-14), which had to write `rest: Rest.None` in every `Command` literal; std-terminal's `Color.Default` defaults hit the same gap, and parameter-defaults (D115) uses the same rule for parameter defaults.

## Acceptance

- [ ] Proof: a struct with `kind: Kind = Kind.None` fails on feat/batch-13 with E0106.
- [ ] Fix: a qualified constructor with no fields, and a qualified constructor call with constant arguments (`Kind.Sized(4)`), are constants for field defaults and every other use of `isConstantValue`.
- [ ] Sample with field defaults that are qualified constructors, output checked; an error test for a non-constant argument (`Kind.Sized(count())`).
