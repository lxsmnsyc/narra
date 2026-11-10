---
id: perceus-reuse
title: "Drop specialization and in-place reuse of unique enum payloads"
status: done
priority: should
type: feature
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - testdata/perceus_reuse_*
  - roadmaps/selfhost/nodes/perceus-reuse/**
shared:
  - selfhost/src/emitstate.nra
  - selfhost/src/emitmatch.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitdecls.nra
  - DECISIONS.md
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: option-inline
    kind: hard
    why: reuse applies only to the heap payloads that remain after the layout change
---

# Drop specialization and in-place reuse of unique enum payloads

## Goal

The second Perceus step, for the recursive enums that still have a heap payload (such as the compiler's AST and type trees). When a `match` arm takes apart a payload that dies there and builds a payload of the same size, and the old payload's count is 1, write the new fields into the old cell instead of a free and a new allocation. When the constructor is known in an arm, release the old value with code for that constructor only, so retains of its fields and the later release cancel out.

## Context

- The perceus outcome describes the ownership and last-use rules this builds on. Read it first.
- Payload layout and `drop_<id>` are in `emitdecls.nra` (see the option-inline outcome for what changed).
- Reuse must keep the D55 rule that a shared payload is never mutated: the uniqueness check (`rc == 1`) is what makes the write safe, and a shared payload falls back to a new allocation.

## Acceptance

- [ ] Samples `testdata/perceus_reuse_*`: a list map over a unique list (reused), the same over a shared list (copied, the original intact), and a tree rewrite.
- [ ] `bench/run.sh` before and after, including the compiler row; the outcome says which programs gained.
- [ ] `sh scripts/test.sh` and `sh scripts/linux-check.sh` pass; the seed is regenerated and the fixed point holds.

## Sizing

L: uncertainty medium to high (reuse analysis is new here), blast radius high. If the gain on the compiler row is within noise after drop specialization, stop there and say so in the outcome.
