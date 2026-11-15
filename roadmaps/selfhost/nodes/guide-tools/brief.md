---
id: guide-tools
title: "Guide chapter: the narrac tools"
status: todo
priority: could
type: docs
story: batch-16
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
verify: sh scripts/test.sh --suite guide
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/guide-tools/**
  - docs/guide/12-tools.md
shared:
  - docs/guide/README.md
  - docs/guide/12-narra-for-typescript-programmers.md
  - docs/guide/01-getting-started.md
  - selfhost/tests/**
depends_on: []
reserves: []
provides: []
---
# Guide chapter: the narrac tools

## Goal

The guide teaches the language but not the tools around it. Add chapter 12, "Tools", and move the TypeScript comparison to chapter 13. The chapter covers `narrac check` and its JSON output, `narrac explain`, `narrac apply-fix`, `narrac fmt` (and its config file), `narrac query`, `narrac test`, and starting `narrac lsp` from an editor. narra-skills and the website both point at this chapter.

## Context

- Commands and flags: `selfhost/src/main.nra`, `driver.nra`, `fmt*.nra`, `query*.nra`, `lsp*.nra`, and `compiler/README.md`.
- The guide README's "Planned" list is out of date: async, threads, SIMD and `@const` functions are built. Fix that list in this node, keeping only what is really not built (package dependencies, and anything else the code shows as missing).
- Learner docs rule: no SPEC or decision-number citations, no assumed TypeScript knowledge.
- Every sample runs in the guide suite. Shell sessions use `sh` fences and are not run.

## Acceptance

- [ ] `docs/guide/12-tools.md` exists, the comparison chapter is 13, and every link in the guide still resolves.
- [ ] The README chapter list and "Planned" list match the compiler.
- [ ] `sh scripts/test.sh --suite guide` passes.

## Sizing

M: one chapter of about the size of chapter 10, plus a renumber.
