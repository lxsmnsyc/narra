---
id: inline-benchmark
title: "Spike: is @inline worth keeping"
status: done
priority: must
type: spike
story: batch-15
early: false
outline: false
milestone: p3-review
target:
ship: auto
land: auto
from:
size: S
spike: true
hypothesis:
verify: sh scripts/test.sh
exclusive: [bench]
fulfilled_by:
gate: true
added: 2026-11-14
spent: 85153
owns:
  - roadmaps/selfhost/nodes/inline-benchmark/**
  - bench/inline/**
shared:
reserves: []
provides: []
depends_on: []
---
# Spike: is @inline worth keeping

## Goal

D119: measure what `@inline` buys. Build the bench programs and `narrac emit selfhost` with and without the attribute (a temporary compiler flag or a branch that ignores it, not committed), on a quiet machine (load under 4; wait and recheck if not).

## Context

- Bench harness: `bench/` and `bench/BASELINE.md`. Use instruction counts and cycles, at least 5 runs each, report spread.
- Cover std's `Options.map`, `andThen`, `orElse` and the compiler's hot `@inline` helpers (`grep -rn '@inline' selfhost/src std/src | wc -l`).
- Clang at -O2 may inline the same calls anyway: that is the question.

## Acceptance

- [ ] `outcome.md`: a table per program, gain with spread, and a recommendation: keep, or remove with a follow-up outline node brief. A gain inside the noise counts as none.
- [ ] Gate: the orchestrator hands the result to the maintainer before any removal.
- [ ] No source change lands from this node.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
