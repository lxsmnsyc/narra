---
id: guide-basics-gaps
title: "Chapter 2: divisions, maps, strings and loops"
status: done
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
  - roadmaps/selfhost/nodes/guide-basics-gaps/**
  - docs/guide/02-values-and-types.md
shared:
  - docs/guide/06-option-and-result.md
reserves: []
provides: []
depends_on:
  - on: guide-corrections
    kind: hard
    why: both edit chapter 2
---
# Chapter 2: divisions, maps, strings and loops

## Goal

Audit rows 115, 117, 127, 10, 120, 33, 34, 5, 36, 45, 46, 79 in `docs/guide/02-values-and-types.md`: divisor proofs and what a check proves (move the rule here from chapter 6, which keeps a link); `Maps.update` (`counts[key] += 1` does not compile); string `+=` and the string builder; C-style `for` and labels (now only in chapter 20); line continuation and multi-line `? :`; `Int.pow`, `floorDiv`, `divmod`; the common `Arrays` and `Maps` functions.

## Rules

- The rows named above are in `roadmaps/selfhost/nodes/docs-concept-audit/audit.md` (inventory, with the source of each concept). Read the source of each concept (`SPEC.md` section, std source, `testdata/` programs) before writing; check every claim with a compiler built from this branch.
- The guide is standalone for learners: no SPEC or decision-record citations, no assumed TypeScript outside chapter 20. Every concept gets a short runnable example, in the fence styles the other chapters use (`narra`, `narra error`, `narra file ...`, with an `output` block where the chapter shows one). The `guide` suite runs the samples.
- Write by the agent-skills:plain-writing rules; run its checker on the files you change.
- Keep the chapter's voice and running example where it has one. Do not restate what another chapter teaches; link to it.

## Acceptance

- [ ] Each concept listed is taught with a runnable example, or a note in the outcome says why not.
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes.
