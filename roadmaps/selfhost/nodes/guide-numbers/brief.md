---
id: guide-numbers
title: "Numbers: sized integers, floats, bits"
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
  - roadmaps/selfhost/nodes/guide-numbers/**
  - docs/guide/02-values-and-types.md
shared:
reserves: []
provides: []
depends_on:
  - on: guide-basics-gaps
    kind: hard
    why: both edit chapter 2
---
# Numbers: sized integers, floats, bits

## Goal

Audit rows 81, 82, 44, 116, 118, 66, 68, 69: sized integers and conversions, `f32`, bitwise operators and shift proofs, float text, `std/math`, `std/random`, `std/hash`. A new section at the end of `docs/guide/02-values-and-types.md`.

## Rules

- The rows named above are in `roadmaps/selfhost/nodes/docs-concept-audit/audit.md` (inventory, with the source of each concept). Read the source of each concept (`SPEC.md` section, std source, `testdata/` programs) before writing; check every claim with a compiler built from this branch.
- The guide is standalone for learners: no SPEC or decision-record citations, no assumed TypeScript outside chapter 20. Every concept gets a short runnable example, in the fence styles the other chapters use (`narra`, `narra error`, `narra file ...`, with an `output` block where the chapter shows one). The `guide` suite runs the samples.
- Write by the agent-skills:plain-writing rules; run its checker on the files you change.
- Keep the chapter's voice and running example where it has one. Do not restate what another chapter teaches; link to it.

## Acceptance

- [ ] Each concept listed is taught with a runnable example, or a note in the outcome says why not.
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes.
