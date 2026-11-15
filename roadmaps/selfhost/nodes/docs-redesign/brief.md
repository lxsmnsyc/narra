---
id: docs-redesign
title: "Docs pages on the new design"
status: todo
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
  - roadmaps/selfhost/nodes/docs-redesign/**
  - website/src/components/DocPage.brm
  - website/src/routes/guide/**
  - website/src/routes/install.brm
  - website/src/routes/spec.brm
  - website/src/docs/**
shared:
  - website/src/styles.css
  - website/src/client.ts
reserves: []
provides: []
depends_on:
  - on: site-design
    kind: hard
    why: uses its tokens and edits the same stylesheet
---
# Docs pages on the new design

## Goal

The guide, install and spec pages on the design from site-design: reading width and type, sidebar with the current chapter, table of contents that follows the scroll, code block chrome (language, copy button with feedback), a pager with motion between chapters, mobile menu. Fix what looks broken on these pages today.

## Acceptance

- [ ] Guide chapter, install and spec pages checked in Chrome, dark and light, phone width.
- [ ] Copy button on every code block, with keyboard access and feedback.
