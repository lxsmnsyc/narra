---
id: or-pattern-crash
title: "Compiler crash on an or-pattern that binds inside nested constructors"
status: todo
priority: should
type: fix
story: batch-4
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - testdata/or_pattern_nested_bind*
  - roadmaps/selfhost/nodes/or-pattern-crash/**
shared:
  - selfhost/src/emitmatch.nra
  - selfhost/src/match.nra
  - selfhost/src/patterns.nra
reserves: []
provides: []
depends_on:
  - on: perceus
    kind: hard
    why: perceus changes the binding code in emitmatch.nra; the fix lands on top of it
---

# Compiler crash on an or-pattern that binds inside nested constructors

## Goal

`narrac` segfaults (exit 139) on an or-pattern whose alternatives bind the same name at different nested positions:

```
match (wrap) {
  Wrap.Pair(Chain.Link(n, _), _) | Wrap.Pair(_, Chain.Link(n, _)) => print(n)
  _ => {}
}
```

The full program is `repro.nra` in this folder. Find where the compiler crashes and fix it so the program compiles and prints the bound value from whichever alternative matched.

## Context

- Found by the perceus fresh review. It crashes on main before batch-4 (commit 4518a91) as well, so it is not caused by perceus or nested-or-guards.
- Proof first: the repro exits 139 on feat/batch-4.

## Acceptance

- [ ] `testdata/or_pattern_nested_bind.nra` covers the repro, the same with a guard, and alternatives that bind two names at different depths; output checked, clean under AddressSanitizer and the leak check.
- [ ] `sh scripts/test.sh` passes.

## Sizing

S: one crash with a minimal repro; may grow if the binding layout of or-patterns needs rework, then say so in the outcome.
