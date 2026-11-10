---
id: closure-arg-temps
title: "Block-bodied arrow arguments do not count as writers"
status: active
priority: should
type: fix
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
verify: sh scripts/test.sh --suite programs
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/closure-arg-temps/**
shared:
  - selfhost/src/emitexprs.nra
  - roadmaps/selfhost/inbox.md
reserves: []
provides: []
depends_on: []
---

# Block-bodied arrow arguments do not count as writers

## Goal

`mayWrite` (`selfhost/src/emitexprs.nra:563`) counts a block-bodied arrow argument as a writer, so an earlier argument in the same call gets a temporary it does not need (arg-eval-order, D-number in its outcome). Building a closure runs none of its body while the arguments are evaluated, the same reasoning that already makes closure locals non-writers. Return false for arrow expressions of either body form.

## Context

- Inbox entry 2026-11-08, from PR #100. The note says the change removes most of the extra C in 5 samples.
- Evidence to record: the emitted C for one of those samples before and after (a temporary that disappears).

## Acceptance

- [ ] A sample where an earlier argument is read and a later argument is a block-bodied arrow that writes the same variable when called later; output unchanged, emitted C has no temporary for the earlier argument.
- [ ] `sh scripts/test.sh` passes; inbox entry removed.

## Sizing

S: one function, one sample.
