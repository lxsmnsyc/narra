---
id: guide-final-pass
title: "Guide index, cross-links and a read-through"
status: todo
priority: could
type: docs
story: batch-21
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/guide-final-pass/**
  - docs/guide/README.md
shared:
  - docs/guide/**
  - skills/narra/**
reserves: []
provides: []
depends_on:
  - on: guide-corrections
    kind: hard
    why: reads the finished chapters
  - on: guide-basics-gaps
    kind: hard
    why: reads the finished chapters
  - on: guide-numbers
    kind: hard
    why: reads the finished chapters
  - on: guide-system-io
    kind: hard
    why: reads the finished chapters
  - on: guide-cell-and-state
    kind: hard
    why: reads the finished chapters
  - on: guide-generators
    kind: hard
    why: reads the finished chapters
  - on: guide-async
    kind: hard
    why: reads the finished chapters
  - on: guide-threads
    kind: hard
    why: reads the finished chapters
  - on: guide-c-interop
    kind: hard
    why: reads the finished chapters
  - on: guide-types-more
    kind: hard
    why: reads the finished chapters
  - on: guide-serialization
    kind: hard
    why: reads the finished chapters
  - on: guide-patterns-more
    kind: hard
    why: reads the finished chapters
  - on: guide-proofs
    kind: hard
    why: reads the finished chapters
  - on: guide-compile-time
    kind: hard
    why: reads the finished chapters
  - on: guide-tools-gaps
    kind: hard
    why: reads the finished chapters
  - on: guide-targets
    kind: hard
    why: reads the finished chapters
  - on: grammar-fixes
    kind: hard
    why: reads the finished chapters
---
# Guide index, cross-links and a read-through

## Goal

After every chapter writer of batch-21 lands, make the guide read as one book:

- `docs/guide/README.md`: the chapter list with one line per chapter as written, and the "built but not taught" and "Planned" lists updated (fixed-size arrays, `ensures`, refinement types and others are now taught).
- Cross-links the writers left as placeholders or one-sentence mentions now that the target chapter exists: chapter 14 names `Cell` (chapter 9 teaches it); chapter 15 leaves `Dynamic<Trait>` out of the non-`Sendable` list (chapter 7 may teach it now); chapters 9 and 14 link forward to chapters written later. Search every chapter for "chapter N" mentions and `Next:` lines and check each target.
- Read chapters 1 to 20 in order as a learner, and fix concepts used before they are taught (move the explanation or add a forward link), contradictions between chapters, and repeated explanations (keep one, link from the others). Keep every chapter's samples running.
- `skills/narra/SKILL.md`: update anything the guide now says differently.

## Acceptance

- [ ] README lists match the chapters; no link points at a missing section.
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes.
