---
id: thread-scope-void
title: "Thread.scope with a body that returns nothing"
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
  - roadmaps/selfhost/nodes/thread-scope-void/**
shared:
  - std/src/thread.nra
  - selfhost/src/check/**
  - selfhost/src/emit/**
  - testdata/**
  - selfhost/tests/src/**
  - docs/guide/15-threads-and-shared-state.md
  - SPEC.md
reserves: []
provides: []
depends_on: []
---
# `Thread.scope` with a body that returns nothing

## Goal

Found by spec-corrections, confirmed by the orchestrator on main (2026-11-20). This program fails with an error inside the standard library, `std/src/thread.nra:42:3: this expression has no value`:

```narra
import * as Thread from "std/thread"
export function run(args: Array<string>): int uses io, alloc {
  Thread.scope((scope) => {
    const _worker = Thread.start(scope, () => 1)
  })
  print("done")
  return 0
}
```

`Thread.scope` is generic in the body's result `T`, and with `T = void` the std body breaks. A scope whose body returns nothing is the common case (start threads, let the scope join them), and SPEC states no restriction. A user must never see an error located in std. Fix it so the program runs (std, or the checker's void handling for generic results, whichever is the cause; the inbox has a 2026-11-14 note that void type arguments work only through special cases in `std/thread` and `std/async`). Then let SPEC section 8 and chapter 15 use a scope with no result where that reads better.

## Acceptance

- [ ] The program above fails on the batch base and runs on the branch; a case covers it.
- [ ] No error from a user program is reported at a location inside `std/src` for this shape.
- [ ] `sh scripts/test.sh` passes.
