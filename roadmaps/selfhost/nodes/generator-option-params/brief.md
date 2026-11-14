---
id: generator-option-params
title: "Generator types with effects nested in Option and results"
status: todo
priority: must
type: fix
story: batch-15
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/generator-option-params/**
  - testdata/generator_effect_types*
  - testdata/errors/generator_effect_types*
shared:
  - selfhost/src/*.nra
  - std/src/iterator.nra
reserves: []
provides: []
depends_on: []
---
# Generator types with effects nested in Option and results

## Goal

Two checker gaps with effect lists on generator types:

1. A bound `Option<Generator<T> uses (E, alloc)>` cannot be passed to any parameter (E0119 whatever the parameter's uses list), and the type prints without its effects in messages.
2. A plain function that passes on a `Generator<T> uses E` parameter cannot write a result type that matches it ("expected Generator<int>, found Generator<int> uses alloc"). `AsyncGenerator` inherits it.

## Context

- Type equality and printing of function and generator types with effect lists, and effect-variable binding through nested types (batch-13 fixed binding through `Option<Generator<T> uses E>` results; this is the parameter and pass-through side).
- After the fix, std/iterator's `chunk` and `window` samples can pass the Option without unwrapping where it is made: simplify them.

## Acceptance

- [ ] Tests fail on the base and pass for both cases, with `AsyncGenerator` too, and the printed type in an error message shows its effects.
- [ ] `sh scripts/test.sh` passes.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
