---
id: effect-variable-values
title: "Functions with effect variables used as values keep their effects"
status: todo
priority: should
type: fix
story: batch-2
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-09
spent:
owns:
  - testdata/errors/effect_variable_value*
  - testdata/effect_variable_value*
  - roadmaps/selfhost/nodes/effect-variable-values/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [decision D91]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes go into the Narra compiler only, after the bootstrap fixed point
---

# Functions with effect variables used as values keep their effects

## Goal

Soundness hole found in PR #105. A function with an effect variable used as a value (`const apply = twice`, with `twice<effect E>(f: (v: int) => int uses E, v: int): int uses E`) makes `apply(loud, 1)` count as pure, through a top-level or a local constant. The store into the parameter is skipped because the binder is another function, and `concreteEffects` drops the variable.

Pick one fix, and record the choice as D91 in DECISIONS.md with the reason:

1. Bind effect variables on calls through values, so `apply(loud, 1)` charges `loud`'s effects.
2. Reject a function with effect variables used as a value (new error with a hint to wrap it in a closure).

Prefer (1) when it fits the existing effect substitution from effects-indirect-calls (PR #98); otherwise (2). The gate decision is the orchestrator's.

## Acceptance

- [ ] Proof on the base: an error test where a pure function calls `apply(loud, 1)` through a top-level constant and through a local constant checks clean before the change.
- [ ] After the change it is rejected (option 1: missing effect; option 2: the new error), with tests in `testdata/errors/effect_variable_value*`.
- [ ] Calls that were correct before still check (a sample that passes a pure function through the value).
- [ ] D91 in DECISIONS.md; SPEC.md section 7 matches.
- [ ] `sh scripts/test.sh` passes.
