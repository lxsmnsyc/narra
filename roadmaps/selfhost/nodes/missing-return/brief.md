---
id: missing-return
title: "Missing return is a compile error"
status: done
priority: should
type: feature
story: batch-3
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
gate: false
added: 2026-11-09
spent: 118305
owns:
  - testdata/missing_return*
  - testdata/errors/missing_return*
  - roadmaps/selfhost/nodes/missing-return/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [code E0969, code E0970]
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 compiler
---

# Missing return is a compile error

## Goal

D92: a function with a result must return on every path. Today `function f(x: int): int { if (x > 0) { return 1 } }` checks clean and `f(0)` panics at runtime ("function ended without returning a value").

A path ends at: `return`; a call to `panic`; an `if`/`else` whose branches all end; an exhaustive `match` whose arms all end; `while (true)` or `for (;;)` with no `break` that leaves it. Any other end of a function with a result is an error at the closing brace, naming the path that falls through. Arrow functions with block bodies and generator bodies follow the same rule where they have a result. The emitted runtime check stays, worded as an internal compiler error. Fix every site the rule rejects in `selfhost/src`, `std/src` and the tests.

## Acceptance

- [ ] Proof on the base: the sample above checks clean and panics.
- [ ] Error tests (E0969, E0970 if a second code is needed) for: a missing `else`, a non-exhaustive path through `match`, a loop with a `break`, a block arrow.
- [ ] Samples for each ending form that is accepted, including `while (true)` without `break` and a final `panic`.
- [ ] SPEC.md states the rule; D92 drops "Not built yet" for this part.
- [ ] `sh scripts/test.sh` passes.
