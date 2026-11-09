---
id: shared-write-sweep
title: "Fix the shared writes in the compiler"
status: todo
priority: should
type: refactor
story:
early: false
outline: true
milestone: m12-spec-gaps
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: shared-write-check
    kind: hard
    why: needs the warnings and their fixes
---

# Fix the shared writes in the compiler

## Goal

Apply the `copy` and `from` fixes across selfhost/src, std/src and the tests (about 120 to 220 edits, 30 to 50 marks), and edit the save/restore sites with a mutable call between save and reset by hand (O9). No behavior change except removed hidden copies.

## Notes

Outline, from D94 and `roadmaps/selfhost/debate/value-semantics/debate.md` (step 4 of 6). Refine with the debate's claims named here.
