---
id: docs-structure-review
title: "Review the structure of the docs"
status: done
priority: could
type: spike
story: batch-20
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: M
spike: true
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/docs-structure-review/**
shared: []
reserves: []
provides: []
depends_on:
  - on: batch-17
    kind: hard
    why: reviews the docs and site as batch-17 leaves them
---
# Review the structure of the docs

## Goal

Maintainer, 2026-11-16: review the docs structure. Some things are missing (perhaps a std API reference), outdated (the install step), or probably unnecessary.

Audit `docs/guide/*`, `docs/install.md`, `SPEC.md` as published, the README and the site navigation, and propose a structure for the maintainer to approve. Known leads, to check, not proof:

- Missing: a std API reference. `narrac query interface MODULE` already prints the exported signatures of a module, so a reference could be generated from `std/src` at build time instead of written by hand. Other candidates: a CLI reference page, a `narra.toml` reference, an error code index (`narrac explain`).
- Outdated: chapter 1 (`docs/guide/01-getting-started.md`, lines 10 to 15) still installs with `git clone` and `sh selfhost/bootstrap.sh`; it should use the one-liner from contract install-command and send building from source to `docs/install.md`. Search the rest for the same.
- Possibly unnecessary: pages or chapters that repeat each other (chapter 11 against a generated std reference, chapter 12 against a CLI reference), internals a learner does not need, and SPEC sections a reader of the site does not use.

Base each proposal on evidence (what the page holds, who reads it, what links to it), not taste. The guide stays standalone (no SPEC or decision references, no assumed TypeScript outside chapter 13).

## Acceptance

- [x] A proposed structure (pages, order, what moves, what is generated, what is dropped) with the reason for each change; the maintainer approves.
- [x] Follow-up nodes for the approved changes.

Approved by the maintainer on 2026-11-20, all changes. Follow-ups: docs-release-pass (batch-20) and nine outline nodes after first-release. LICENSE: MIT, in docs-release-pass.
