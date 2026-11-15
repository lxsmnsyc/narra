---
id: docs-search
title: "Search for the docs"
status: active
priority: could
type: feature
story: batch-17
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
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/docs-search/**
  - website/src/search/**
  - website/plugins/search.ts
  - website/src/components/Search.brm
shared:
  - website/src/routes/_layout.brm
  - website/src/styles.css
  - website/src/theme.ts
  - website/vite.config.ts
  - website/package.json
  - website/pnpm-lock.yaml
  - website/src/env.d.ts
  - website/plugins/markdown.ts
reserves: []
provides: []
depends_on: []
---
# Search for the docs

## Goal

Maintainer, 2026-11-16: the docs miss a search bar like the maintainer's other sites. Follow those sites (`../seroval/website`: `src/components/Search.tsx`, `createSearchEntries` in `plugins/content.ts`, the `.search-*` rules in `src/styles.css`): one minisearch entry per section (page title, heading, text) built at build time into `virtual:search-index`; the index and minisearch load the first time search opens; a trigger in the header with a `⌘K` hint; Cmd/Ctrl+K and `/` open a command palette; results show the page or section, a snippet, arrow keys move, Enter opens, Esc closes. Index the guide chapters, the install page and the spec.

The site is brumate, not Solid: write the palette as a `.brm` component with signals, and follow ui-style (focus trap, focus returns to the trigger, reversible motion, reduced motion).

## Acceptance

- [ ] Header trigger and Cmd/Ctrl+K and `/` open search on every page.
- [ ] A query such as `mutable` or `uses alloc` finds the right sections; Enter opens the section anchor.
- [ ] Index size and the extra page weight recorded in outcome.md; nothing loads before search opens.
