---
id: shared-write-sweep-rest
title: "Fix the shared writes in the rest of the compiler and the tests"
status: done
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/shared-write-sweep-rest/**
shared:
  - selfhost/src/**
  - selfhost/narra.toml
  - selfhost/tests/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: shared-write-sweep
    kind: hard
    why: both edit selfhost/src; the debate runs D94 one node at a time, and the switch comes off only when every module is clean
---

# Fix the shared writes in the rest of the compiler and the tests

## Goal

D94 step 4, second half. Split from shared-write-sweep on 2026-11-11 (see that brief). This node covers the files shared-write-sweep leaves out: `emit*.nra`, `cnames.nra`, `fmt*.nra`, `query*.nra`, `lsp*.nra`, `trivia*.nra`, `parser.nra`, `lexer.nra`, `modules.nra`, `manifest.nra` and `driver.nra` in `selfhost/src` (about 58 of the 198 warnings), plus `selfhost/tests/src` (3) and `testdata` (36).

- The same fixes as shared-write-sweep: `narrac apply-fix` for `copy(...)` and `from`, hand edits for loop-held writes.
- `testdata`: the 36 programs that warn are copy samples that keep a hidden copy on purpose (`copy_report`, `loop_unique_hoist`, `copy_hoist_*`, `mutable_for*`, ...). Keep what each sample tests. Where the hidden copy is the point, move the sample to `testdata/errors/` with the E0978 golden, or make the copy visible and update its `.copies` file. Where it is not, apply the fix.
- Remove `shared-results = "unchecked"` from `selfhost/narra.toml`. If no other package uses the switch, remove it (the list of places is in `nodes/from-marker/outcome.md`, Deviations).

## Acceptance

- [ ] `narrac check selfhost` reports no E0976 and no E0978, without the switch.
- [ ] The same for `std/src`, `selfhost/tests` and every `testdata` program that is not an error sample.
- [ ] `sh selfhost/bootstrap.sh` reaches the fixed point.
- [ ] Instructions retired for `narrac emit selfhost` before and after, in the outcome.
- [ ] `sh scripts/test.sh` passes.
