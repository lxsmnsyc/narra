---
id: compiler-std-sorts
title: "The compiler uses std sort and char classes"
status: todo
priority: should
type: refactor
story: batch-23
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/compiler-std-sorts/**
shared:
  - selfhost/src/**
  - testdata/**
reserves: []
provides: []
depends_on: []
---
# The compiler uses std sort and char classes

## Goal

Batch-22 added `Arrays.sort`, `Arrays.sortBy` (stable merge sort, in place) and ASCII `std/char` classes. The committed seed now compiles them, so `selfhost/src` may use them. Replace the compiler's own copies:

- The 19 hand-written sorts (`grep -rn -i -e 'insertion sort' -e 'function sort' selfhost/src`): `sortStrings` in `tools/querygraph.nra` and `driver/modules.nra`, `sortNames` in `driver/workspace.nra`, `sortDiagnostics` in `driver/driver.nra`, `sortAuditEntries` in `tools/audit.nra`, the sorts in `driver/fixes.nra` and `driver/lockfile.nra`, and the rest the grep finds. Where the order must stay stable (the comments say so), `sortBy` is stable.
- Character checks: `isLetter`, `isDigit`, `isHexDigit` in `syntax/lexer.nra`, `isSpace` in `syntax/trivia.nra`, `isLetter` in `driver/manifest.nra`. Use `std/char` where its definition matches exactly (ASCII only, same set of characters). The lexer takes bytes as `int`: if the std function needs a `char` and the conversion costs time in the lexer, keep the lexer's own function and say why in the outcome.

## Rules

- Behavior does not change: every golden and emitted C output stays the same, except where a sort was not stable and its order was visible; say so in the outcome if that happens.
- The compiler must not get slower: compare `bench/run.sh -n 5 compiler` before and after, and report both with the load average.
- Delete a helper once nothing calls it. No function-level comments, no `<noun>Of` names, no D-numbers in code.

## Acceptance

- [ ] No hand-written sort left in `selfhost/src`, or each one kept has a reason in the outcome.
- [ ] Character checks use `std/char` where the definitions match.
- [ ] `sh scripts/test.sh` passes; `sh selfhost/bootstrap.sh` reaches the fixed point; the compiler row is not slower beyond noise.
