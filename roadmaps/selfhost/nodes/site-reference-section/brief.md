---
id: site-reference-section
title: "Reference section on the site"
status: done
priority: could
type: feature
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
  - roadmaps/selfhost/nodes/site-reference-section/**
  - docs/reference/README.md
shared:
  - website/src/**
  - website/plugins/**
  - website/vercel.json
reserves: []
provides: []
depends_on: []
---
# Reference section on the site

## Goal

`/reference` index, header link (Install, Guide, Reference), routes in `website/src/pages.ts`, search sources in `website/plugins/search.ts`, `docs/reference` in the `ignoreCommand` of `website/vercel.json`. Proposal change 9.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refined 2026-11-20 into batch-21 (docs-concept-audit, approved).

## Acceptance

- [ ] `/reference` lists the reference pages (Std, Commands, narra.toml, Errors, Spec); the header shows Install, Guide, Reference; search indexes `docs/reference`.
- [ ] `pnpm --dir website build` passes.
