---
id: const-function-effects
title: "A constant holding a function with effects"
status: done
priority: could
type: fix
story: batch-21
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/const-function-effects/**
shared:
  - selfhost/src/check/**
  - testdata/**
  - selfhost/tests/src/**
  - docs/guide/04-structs-and-enums.md
reserves: []
provides: []
depends_on: []
---
# A constant holding a function with effects

## Goal

Found by guide-types-more, confirmed by the orchestrator on main (2026-11-20): a top-level `const` with no written type cannot hold a function that has effects.

```narra
import * as Int from "std/int"
function describe(count: int): string uses alloc {
  return `n=${Int.toString(count)}`
}
const shown = describe
export function run(args: Array<string>): int uses io, alloc {
  print(shown(3))
  return 0
}
```

fails with `expected (int) => string, found (int) => string uses alloc` at `const shown = describe`. Writing the type out (`const shown: (int) => string uses alloc = describe`) works, so the inferred type of the constant drops the effects. Fix the inference so the constant's type carries the function's effects (and check `export const` re-exports of a function from another module, the case the guide hit). Then let chapter 4's `shop` re-export sample pass on a function with effects if that reads better.

## Acceptance

- [ ] A case with the program above, and one with `export const` of an imported function with effects, fails on the batch base and passes on the branch.
- [ ] `sh scripts/test.sh` passes.
