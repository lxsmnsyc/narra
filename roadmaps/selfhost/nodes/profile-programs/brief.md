---
id: profile-programs
title: "Spike: why sieve and particles trail Go and Rust"
status: active
priority: should
type: spike
story: batch-5
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: M
spike: true
hypothesis:
verify: 
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/profile-programs/**
shared: []
reserves: []
provides: []
depends_on:
  - on: batch-4
    kind: hard
    why: builds on the batch-4 compiler and its measurements
---

# Spike: why sieve and particles trail Go and Rust

## Goal

At the end of batch-4, `sieve` takes 0.265 s against Go 0.177 s and Rust 0.148 s, and `particles` 0.108 s against Go 0.074 s. Find where the extra time goes in the emitted C and propose fixes.

## Context

- Profile both programs built `-release` (`sample` on macOS), and read their emitted C next to the Go and Rust versions in `bench/`. Likely suspects to confirm or rule out: array element updates going through copy-on-write checks, bounds checks in the hot loop, and float or struct copies. Record the numbers against the end-of-batch-4 section of `bench/BASELINE.md`.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [ ] `outcome.md` places the gap in named emitted-C constructs with a profile, and proposes briefs with expected gains.

