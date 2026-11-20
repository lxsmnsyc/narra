---
id: guide-threads
title: "Chapter 15: threads and shared state"
status: active
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
  - roadmaps/selfhost/nodes/guide-threads/**
  - docs/guide/15-threads-and-shared-state.md
shared:
reserves: []
provides: []
depends_on:
  - on: guide-renumber
    kind: hard
    why: the chapter files are renamed and the new chapter files created there
---
# Chapter 15: threads and shared state

## Goal

Audit rows 148 to 151, 140 in `docs/guide/15-threads-and-shared-state.md`: `Thread.scope`, `Parallel.map`, channels, `Shared`, `Atomic`, locks, `Sendable`. Say plainly how threads differ from the async loop of chapter 14 and when to use which.

## Rules

- The rows named above are in `roadmaps/selfhost/nodes/docs-concept-audit/audit.md` (inventory, with the source of each concept). Read the source of each concept (`SPEC.md` section, std source, `testdata/` programs) before writing; check every claim with a compiler built from this branch.
- The guide is standalone for learners: no SPEC or decision-record citations, no assumed TypeScript outside chapter 20. Every concept gets a short runnable example, in the fence styles the other chapters use (`narra`, `narra error`, `narra file ...`, with an `output` block where the chapter shows one). The `guide` suite runs the samples.
- Write by the agent-skills:plain-writing rules; run its checker on the files you change.
- Keep the chapter's voice and running example where it has one. Do not restate what another chapter teaches; link to it.

## Acceptance

- [ ] Each concept listed is taught with a runnable example, or a note in the outcome says why not.
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes.
