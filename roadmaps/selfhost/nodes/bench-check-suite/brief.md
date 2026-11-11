---
id: bench-check-suite
title: "A test suite checks every benchmark program"
status: todo
priority: should
type: chore
story: batch-7
early: false
outline: false
milestone: m6-performance
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
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/bench-check-suite/**
shared:
  - selfhost/tests/src/registry.nra
  - selfhost/tests/src/bench.nra
  - bench/**
reserves: []
provides: []
depends_on: []
---

# A test suite checks every benchmark program

## Goal

No suite builds `bench/*/*.nra`, so proven-operands broke `bench/interp` (`acc % n` without a proof) and only the end-of-batch benchmark found it (wave-19). Add a suite named `bench` that runs `narrac check` on every `bench/<name>/<name>.nra`, one case per program. A failing benchmark fails the run. Suites are Narra modules in `selfhost/tests/src/`, registered in `listSuites()` (`registry.nra`); model it on `leaks.nra`. `scripts/test.sh --suite bench` then picks it up. `bench/**` is listed only in case a benchmark fails the check when the suite first runs.

## Acceptance

- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite <new suite>` passes, and fails when a benchmark has a type error (show it once with a scratch edit, not committed).
- [ ] Fast check passes: `sh scripts/test.sh`
