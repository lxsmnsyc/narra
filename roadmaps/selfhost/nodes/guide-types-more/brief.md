---
id: guide-types-more
title: "Generic types, aliases, typeof, Dynamic"
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
  - roadmaps/selfhost/nodes/guide-types-more/**
  - docs/guide/04-structs-and-enums.md
shared:
  - docs/guide/07-traits.md
reserves: []
provides: []
depends_on:
  - on: guide-renumber
    kind: hard
    why: the chapter files are renamed and the new chapter files created there
---
# Generic types, aliases, typeof, Dynamic

## Goal

Audit rows 87, 103, 107, 53, 91, 92, 109, 111, 108, 94 in `docs/guide/04-structs-and-enums.md` and `docs/guide/07-traits.md`: generic structs and enums and their constraints, generic `implement`, type aliases and re-exporting, `typeof`, `Dynamic<Trait>`, `Hashable`, `Self` in a trait result, distinct trait lists.

## Rules

- The rows named above are in `roadmaps/selfhost/nodes/docs-concept-audit/audit.md` (inventory, with the source of each concept). Read the source of each concept (`SPEC.md` section, std source, `testdata/` programs) before writing; check every claim with a compiler built from this branch.
- The guide is standalone for learners: no SPEC or decision-record citations, no assumed TypeScript outside chapter 20. Every concept gets a short runnable example, in the fence styles the other chapters use (`narra`, `narra error`, `narra file ...`, with an `output` block where the chapter shows one). The `guide` suite runs the samples.
- Write by the agent-skills:plain-writing rules; run its checker on the files you change.
- Keep the chapter's voice and running example where it has one. Do not restate what another chapter teaches; link to it.

## Acceptance

- [ ] Each concept listed is taught with a runnable example, or a note in the outcome says why not.
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes.
