---
id: shared-write-sweep
title: "Fix the shared writes in the compiler"
status: todo
priority: should
type: feature
story: batch-6
early: false
outline: false
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
owns:
  - roadmaps/selfhost/nodes/shared-write-sweep/**
shared:
  - selfhost/src/**
  - selfhost/narra.toml
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: shared-write-check
    kind: hard
    why: needs the warnings and their fixes
  - on: proven-operands
    kind: hard
    why: both rewrite call sites across selfhost/src; the debate orders D94 after the D92 nodes
  - on: std-no-panic
    kind: hard
    why: both rewrite call sites across selfhost/src and std/src; the debate orders D94 after the D92 nodes
---

# Fix the shared writes in the compiler

## Goal

D94 step 4 (debate `roadmaps/selfhost/debate/value-semantics/debate.md`, P19, O9, O10). Make `selfhost/src`, `std/src`, `selfhost/tests/src` and `testdata` clean under both rules, so shared-write-error can turn the warning into an error:

- Shared results (E0976): add `from <parameter>` to the about 150 compiler functions with 185 shared returns, and `copy(...)` where a function shares two parameters (about 15). Use `narrac apply-fix`. Then remove `shared-results = "unchecked"` from `selfhost/narra.toml`, and remove the switch itself if no other package uses it (the list of places is in `nodes/from-marker/outcome.md`, Deviations).
- Shared writes (E0978 warnings): apply the `copy(...)` fixes, then edit by hand the save/restore sites with a `mutable` call between save and reset (O9).

No behavior change except removed hidden copies. The emitted C of the compiler changes; the bootstrap fixed point must still hold.

## Notes

- The seed already reads `copy(...)` and `from` (seed updated at the end of batch-5). No seed update is needed first.
- Size is set from shared-write-check's counts. If the orchestrator split this node, this brief covers only the paths in `owns`.
- Run after proven-operands and std-no-panic, which also rewrite compiler call sites.

## Acceptance

- [ ] `narrac check selfhost` reports no E0976 and no shared-write warning, without the `unchecked` switch.
- [ ] The same for `std/src`, `selfhost/tests` and every `testdata` program that is not an error sample.
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point.
- [ ] Instructions retired for `narrac emit selfhost` before and after, in the outcome (the copies cost something; a large rise is a finding).
- [ ] `sh scripts/test.sh` passes.
