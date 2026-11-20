---
id: guide-corrections
title: "Fix the guide's wrong statements and the std list"
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
size: S
spike: false
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/guide-corrections/**
  - docs/guide/01-getting-started.md
  - docs/guide/08-modules-and-packages.md
  - skills/narra/SKILL.md
shared:
  - docs/guide/02-values-and-types.md
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: guide-renumber
    kind: hard
    why: the chapter files are renamed and the new chapter files created there
---
# Fix the guide's wrong statements and the std list

## Goal

Audit rows 42, 7, 50, 58: chapter 2 says tuples cannot be indexed (`pair[0]` compiles) and that only a `let` can be passed as `mutable` (a `const` compiles); chapters 1 and 8 and `skills/narra/SKILL.md` leave `copy` out of the functions available without an import; chapter 8 lists 14 std modules while the compiler accepts 54 (list them all, grouped, each with one line, until the std reference exists). Also: the `SPEC.md` title still says "(draft)"; drop it, so the site's `/spec` page reads "Narra specification" (maintainer is releasing v0.1.0). Check each against the compiler first; fix the text to match the compiler.

## Rules

- The rows named above are in `roadmaps/selfhost/nodes/docs-concept-audit/audit.md` (inventory, with the source of each concept). Read the source of each concept (`SPEC.md` section, std source, `testdata/` programs) before writing; check every claim with a compiler built from this branch.
- The guide is standalone for learners: no SPEC or decision-record citations, no assumed TypeScript outside chapter 20. Every concept gets a short runnable example, in the fence styles the other chapters use (`narra`, `narra error`, `narra file ...`, with an `output` block where the chapter shows one). The `guide` suite runs the samples.
- Write by the agent-skills:plain-writing rules; run its checker on the files you change.
- Keep the chapter's voice and running example where it has one. Do not restate what another chapter teaches; link to it.

## Acceptance

- [ ] Each concept listed is taught with a runnable example, or a note in the outcome says why not.
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes.
