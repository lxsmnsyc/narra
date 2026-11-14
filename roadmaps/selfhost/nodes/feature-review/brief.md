---
id: feature-review
title: "Feature review: what to remove, what to adopt"
status: todo
priority: should
type: spike
story: batch-13
early: false
outline: false
milestone: p3-review
target:
ship: auto
land: auto
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/debate/feature-review/**
  - roadmaps/selfhost/nodes/feature-review/**
shared: []
reserves: [decision D119]
provides: []
depends_on:
  - on: async-generators-design
    kind: hard
    why: the review weighs the finished p1 language
  - on: threads-design
    kind: hard
    why: the review weighs the finished p1 language
  - on: simd-design
    kind: hard
    why: the review weighs the finished p1 language
  - on: regions-plan
    kind: hard
    why: the review weighs the finished p1 language
  - on: tagged-templates-design
    kind: hard
    why: the review weighs the finished p1 language
  - on: parameter-defaults-design
    kind: hard
    why: the review weighs the finished p1 language
  - on: refinement-types-design
    kind: hard
    why: the review weighs the finished p1 language
  - on: uses-list-design
    kind: hard
    why: the review weighs the finished p1 language

---

# Feature review: what to remove, what to adopt

## Goal

Review the whole language in SPEC.md against what the self-hosted compiler and std actually use: features that can be removed or simplified before the language is released, and features worth adopting from other languages. Run it as a goal debate (agent-skills:goal-debate), with evidence from the compiler and std sources. Ends with a list of proposals for the maintainer, each to become a node or a decision record.

## Acceptance

- [ ] `roadmaps/selfhost/debate/feature-review/debate.md` lists each proposal (remove, simplify or adopt) with evidence from SPEC, the compiler and std sources, its cost and its value.
- [ ] D119 records the proposals the maintainer accepts; each accepted one has an outline node in a later phase or batch.

## Notes

Outline, added 2026-11-13 when the maintainer set the phase order (syntax, std, feature review, packages, distribution, docs).
