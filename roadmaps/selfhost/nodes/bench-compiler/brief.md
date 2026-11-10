---
id: bench-compiler
title: "Benchmark the compiler compiling itself"
status: todo
priority: should
type: chore
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: bench/run.sh -n 3 compiler
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - bench/run.sh
  - bench/README.md
  - bench/BASELINE.md
  - roadmaps/selfhost/nodes/bench-compiler/**
shared: []
reserves: []
provides: []
depends_on:
  - on: bench-harness
    kind: hard
    why: extends bench/run.sh
---

# Benchmark the compiler compiling itself

## Goal

Add a `compiler` row to `bench/run.sh`: the median wall time of `narrac emit selfhost` (the release-built compiler emitting its own C), plus peak memory where the platform reports it (`/usr/bin/time -l` on macOS, `-v` on Linux). Then re-record `bench/BASELINE.md` on an idle machine for all six programs and the compiler. This is the baseline perceus, option-inline and perceus-reuse report against.

## Context

- There is no compiler benchmark today: `narrac test -time` times tests only, and `selfhost/bootstrap.sh` prints size and line counts, not time.
- `bench/CLAUDE.md`: the current baseline was recorded on a busy machine; re-record it on an idle one.
- The compiler row has no Go or Rust column; print `-` there.

## Acceptance

- [ ] `bench/run.sh compiler` prints a median time and, where available, peak memory.
- [ ] `bench/BASELINE.md` has a new dated section with all seven rows, the machine, and how idle it was.
- [ ] `bench/README.md` describes the compiler row.

## Sizing

S: one script and two docs, existing pattern.
