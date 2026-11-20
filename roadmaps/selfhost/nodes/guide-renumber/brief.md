---
id: guide-renumber
title: "Renumber the guide to the 20-chapter order"
status: done
priority: could
type: chore
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
  - roadmaps/selfhost/nodes/guide-renumber/**
  - docs/guide/**
shared:
  - docs/install.md
  - website/src/**
  - website/plugins/**
  - skills/narra/**
  - README.md
  - selfhost/tests/src/guide.nra
reserves: []
provides: []
depends_on: []
---
# Renumber the guide to the 20-chapter order

## Goal

Apply the chapter order the maintainer approved on 2026-11-20 (`docs-concept-audit/audit.md`, "Chapter order changes"), before any writer starts:

- `git mv` chapter 11 to `13-standard-library.md`, 12 to `18-tools.md`, 13 to `20-narra-for-typescript-programmers.md`.
- Create the new chapter files with their title, a one-paragraph introduction of what the chapter will teach, and nothing else: `11-files-and-the-system.md`, `12-generators.md`, `14-async.md`, `15-threads-and-shared-state.md`, `16-unsafe-code-and-c.md`, `17-proofs-and-compile-time.md`, `19-building-for-other-systems.md`. Titles: Files and the system; Generators; Async; Threads and shared state; Unsafe code and C; Proofs and compile-time constants; Building for other systems.
- Update every link to a renamed chapter (`git grep -n "guide/1[123]-"`, also `docs/install.md`, `website/src/routes/index.brm`, `skills/narra/`, `README.md`), the chapter list in `docs/guide/README.md` (with the new chapters), and the chapter number in each chapter's text where it names another chapter by number.
- The site lists chapters by glob (`website/src/pages.ts`); check the sidebar order and the search index (`website/plugins/search.ts`). The `guide` test suite (`selfhost/tests/src/guide.nra`) reads the chapters; keep it passing.

No content changes beyond links and numbers.

## Acceptance

- [ ] 20 chapter files in the approved order; no link points at an old chapter file (`git grep` shows none).
- [ ] `NARRAC=<built narrac> sh scripts/test.sh --suite guide` passes; `pnpm --dir website build` passes, and the sidebar lists the 20 chapters in order.
