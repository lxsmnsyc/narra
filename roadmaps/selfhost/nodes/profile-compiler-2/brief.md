---
id: profile-compiler-2
title: "Spike: profile the compiler again after batch-4"
status: done
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
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent: 281662
owns:
  - roadmaps/selfhost/nodes/profile-compiler-2/**
shared: []
reserves: []
provides: []
depends_on:
  - on: batch-4
    kind: hard
    why: builds on the batch-4 compiler and its measurements
---

# Spike: profile the compiler again after batch-4

## Goal

Batch-4 cut `narrac emit selfhost` from 9.30G to 2.59G instructions, so the shares the first profile found (allocation 40%, strings 20%) no longer hold. Profile the compiler at the end of batch-4 the same way and propose the next nodes, most valuable first.

## Context

- Reuse the scripts in `roadmaps/selfhost/nodes/profile-compiler/` (`build.sh`, `record.sh`, `analyze.py`). Compare each category and phase with the first profile, and check the follow-ups below against the new numbers before they start.
- Source: the batch-4 final report (2026-11-10) and the batch-4 inbox entries.

## Acceptance

- [ ] `outcome.md` has the same tables as the first profile at a named commit, and a ranked list of proposed briefs. It says for each of tail-loop-borrow, cache-type-names-rest and string-concat-one-alloc whether the profile still supports it.

