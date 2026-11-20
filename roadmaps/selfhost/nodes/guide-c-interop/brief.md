---
id: guide-c-interop
title: "Chapter 16: unsafe code and C"
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
  - roadmaps/selfhost/nodes/guide-c-interop/**
  - docs/guide/16-unsafe-code-and-c.md
shared:
reserves: []
provides: []
depends_on:
  - on: guide-renumber
    kind: hard
    why: the chapter files are renamed and the new chapter files created there
---
# Chapter 16: unsafe code and C

## Goal

Audit rows 142, 159 to 164, 182 in `docs/guide/16-unsafe-code-and-c.md`: the `unsafe` effect and blocks, `@trusted`, `Ptr<T>`, `extern` blocks and C type mapping, `@extern`, linking and `[link] search` in `narra.toml`, the variadic limit, `std/cstring`, `narrac audit`, `@target` on an extern block. E0996 (an extern call outside `unsafe`) is the first error a learner meets. Use a small real C library from the platform (for example `libm` or a C file in the sample) so the example runs in the `guide` suite.

## Rules

- The rows named above are in `roadmaps/selfhost/nodes/docs-concept-audit/audit.md` (inventory, with the source of each concept). Read the source of each concept (`SPEC.md` section, std source, `testdata/` programs) before writing; check every claim with a compiler built from this branch.
- The guide is standalone for learners: no SPEC or decision-record citations, no assumed TypeScript outside chapter 20. Every concept gets a short runnable example, in the fence styles the other chapters use (`narra`, `narra error`, `narra file ...`, with an `output` block where the chapter shows one). The `guide` suite runs the samples.
- Write by the agent-skills:plain-writing rules; run its checker on the files you change.
- Keep the chapter's voice and running example where it has one. Do not restate what another chapter teaches; link to it.

## Acceptance

- [ ] Each concept listed is taught with a runnable example, or a note in the outcome says why not.
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes.
