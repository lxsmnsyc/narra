---
id: decls-quadratic
title: "narrac decls is quadratic in the declarations of a module"
status: todo
priority: should
type: fix
story: batch-3
early: false
outline: false
milestone: m6-performance
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
spent:
owns:
  - testdata/declsdump/perf*
  - roadmaps/selfhost/nodes/decls-quadratic/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the batch-2 compiler
---

# narrac decls is quadratic in the declarations of a module

## Goal

Inbox 2026-10-30 (from the review of #34): `narrac decls` took 0.3 s for 2000 declarations in one module, 1.15 s for 4000 and 4.9 s for 8000, measured on the old Go-era compiler. Not located; likely the declaration passes or the dump.

Prove it first on the current compiler: generate modules of 2000, 4000 and 8000 declarations, time `narrac decls` and `narrac check` (several runs, report the spread), and profile to find the quadratic step. If the growth is linear now, write "Not reproduced" with the numbers and stop. Otherwise fix the cause and show the growth is linear.

## Acceptance

- [ ] Timings on the base and on the branch at 2000, 4000 and 8000 declarations, beyond the noise.
- [ ] The cause named in the outcome (file and function).
- [ ] A test that would catch the quadratic step coming back, if one can be made cheap (for example a count-based check, not a timing).
- [ ] `sh scripts/test.sh` passes.
