---
id: docs-concept-audit
title: "Audit the docs for missing concepts"
status: active
priority: could
type: spike
story: batch-21
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
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/docs-concept-audit/**
shared:
reserves: []
provides: []
depends_on: []
---
# Audit the docs for missing concepts

## Goal

Maintainer, 2026-11-20: "The docs is missing a lot of crucial concepts." docs-structure-review looked at the structure (pages, order, generated references); this node looks at the content: which concepts a Narra user needs and where the docs teach them.

Build the concept inventory from the sources of truth, not from the guide: `SPEC.md`, `DECISIONS.md`, `compiler/README.md` (supported features and differences), the std modules (`std/src`, the builtin registry), the `narrac` commands, the diagnostics with explanations (`narrac explain`), `testdata/` and `samples/` for features in use, and the skill at `skills/narra/` (what an agent is told it must know). Then map each concept to where `docs/guide/*`, `docs/install.md` and the site teach it: taught (with a runnable example), mentioned only, or missing.

Rank each concept: crucial (a user writing a first real program hits it, or the compiler rejects code without it, or it sets Narra apart), important, or niche. Value semantics, ownership and `mutable`, effects (`uses`), `from` and sharing, errors and `Result`, closures and `escaping`, generics, traits, async and threads, packages and workspaces, `@const`, `unsafe` and C interop, testing, and the tools are the likely areas, but the inventory decides.

## Acceptance

- [ ] `audit.md` in this folder: the inventory as a table (concept, source, rank, where the docs teach it, status), then the gaps grouped into proposed writing nodes (id, chapter or page, concepts covered, size), with the crucial gaps first. Reconcile it with the nine outline nodes after first-release (`site-reference-section`, `std-interface-query`, `std-reference`, `std-doc-comments`, `cli-reference`, `manifest-reference`, `error-index`, `spec-decision-links`, `guide-concurrency`): say which become part of this batch.
- [ ] The maintainer approves the list.
