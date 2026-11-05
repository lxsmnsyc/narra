---
id: guide-typescript
title: "Language guide: Narra for TypeScript programmers"
status: todo
priority:
type: docs
story: language-guide
early: false
outline: false
milestone:
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run TestGuide
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-05
spent:
owns:
  - docs/guide/11-*.md
  - roadmaps/selfhost/nodes/guide-typescript/**
shared:
  - docs/guide/README.md
reserves: []
provides: []
depends_on:
  - on: guide-basics
    kind: hard
    why: reads the "For chapter 11" list in its outcome and adds a link to the index it writes
  - on: guide-advanced
    kind: hard
    why: reads the "For chapter 11" list in its outcome
---

# Language guide: Narra for TypeScript programmers

## Goal

Write chapter 11 of `docs/guide/`, a comparison for readers who know TypeScript. Chapters 1 to 10 teach Narra standalone and leave comparisons out (maintainer, 2026-11-05). This chapter collects them in one place: what carries over, what differs and why in practical terms, and for each TypeScript habit the Narra form and a link to the chapter that teaches it.

## Context

- The "For chapter 11" lists in `nodes/guide-basics/outcome.md` and `nodes/guide-advanced/outcome.md` (read them from the story branch `feat/language-guide`).
- The TypeScript-habit diagnostics in `compiler/testdata/diagnostics/ts_*.nra` and `compiler/testdata/errors/ts_*.nra` list the habits the compiler already catches, with their messages.
- Samples follow `contracts/guide-samples.md`.

## Acceptance

- [ ] `docs/guide/11-narra-for-typescript-programmers.md` covers every item of both "For chapter 11" lists and the TypeScript-habit diagnostics, grouped by topic, each with a tested Narra sample or a `narra error` sample.
- [ ] No `SPEC.md`, section or D-number references.
- [ ] `docs/guide/README.md` links the chapter, marked as optional reading for TypeScript programmers.
- [ ] `TestGuide` passes.

## Sizing

S. One prose file and one index line; the content is collected by the two earlier children.
