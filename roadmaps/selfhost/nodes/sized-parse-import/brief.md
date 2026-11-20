---
id: sized-parse-import
title: "Sized parse without std/int fails in the C compiler"
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
  - roadmaps/selfhost/nodes/sized-parse-import/**
shared:
  - selfhost/src/emit/**
  - selfhost/src/check/**
  - selfhost/src/runtime/**
  - testdata/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on: []
---
# Sized parse without std/int fails in the C compiler

## Goal

Found by guide-numbers, confirmed by the orchestrator on main (2026-11-20): this program passes `narrac check` but `narrac build` fails in the C compiler with `call to undeclared function 'narra_sized_parse'`:

```narra
import * as U8 from "std/u8"
export function run(args: Array<string>): int uses io, alloc {
  if (U8.parse("7", 10) is Ok(_value)) {
    print("ok")
  }
  return 0
}
```

Importing `std/int` hides it, so the runtime piece for sized parse is pulled in only through `std/int`. Make every sized module's `parse` (and any other sized function with the same shape) emit what it needs on its own. Check the other sized modules and functions for the same gap.

## Acceptance

- [ ] A case with the program above (and one per sized module, or a table case) fails on the batch base and passes on the branch.
- [ ] `sh scripts/test.sh` passes.
