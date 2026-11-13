---
id: field-default-alloc
title: "A struct literal that leaves out an array field default is not charged alloc"
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
  - roadmaps/selfhost/nodes/field-default-alloc/**
shared:
  - selfhost/src/effects.nra
reserves: []
provides: []
depends_on:
  - on: parameter-defaults
    kind: hard
    why: the fix reuses walkFilledDefaults from parameter-defaults
---

# A struct literal that leaves out an array field default is not charged alloc

## Goal

`struct Bag { items: Array<int> = [1, 2] }` with `return {}` compiles in a function without `uses alloc`, while writing `{ items: [1, 2] }` is rejected. The effect pass does not walk field defaults filled into a struct literal. Found by the parameter-defaults spike and confirmed by its build; the fix can reuse `walkFilledDefaults`.

## Acceptance

- [ ] Proof: the exported function compiles on feat/batch-13 without `uses alloc`.
- [ ] Fix: filled field defaults charge their effects (alloc for array, map and string-building defaults) at the literal.
- [ ] Error test for the missing `alloc`; a sample with the effect declared runs clean.
