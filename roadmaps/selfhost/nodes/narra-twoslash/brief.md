---
id: narra-twoslash
title: "Twoslash for Narra on the website"
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
  - roadmaps/selfhost/nodes/narra-twoslash/**
  - website/plugins/twoslash.ts
  - website/src/twoslash/**
  - website/scripts/narrac.ts
shared:
  - website/plugins/markdown.ts
  - website/src/styles.css
  - website/src/client.ts
  - website/vite.config.ts
  - website/package.json
  - website/vercel.json
reserves: []
provides: []
depends_on:
  - on: query-hovers
    kind: behavior
    via: query-hovers
  - on: docs-highlighting
    kind: hard
    why: both change the fence renderer in plugins/markdown.ts
---
# Twoslash for Narra

## Goal

Port the idea of twoslash (https://shikijs.github.io/twoslash/) to Narra code samples on the site: at build time, run `narrac query hovers --json` on each `narra` fence that compiles as a file, and render hover popups on identifiers and inline error lines under the line they belong to. `narra error` fences show the compiler's real message inline. No client-side compiler: the page ships the popups as HTML plus a small script for placement and keyboard focus.

The build needs narrac. Locally, build it from the seed (`sh selfhost/bootstrap.sh`) once and cache it. On Vercel, the build image may lack a C compiler: find out with docker (`amazonlinux:2023`, the Vercel build image base), and fall back to the wasm narrac from `scripts/wasm-build.sh` if needed. Record the choice. No deploy by an agent.

Popups follow the design tokens of site-design: open on hover and focus, close on Escape, reversible motion, reduced motion respected.

## Acceptance

- [ ] Hover a name in a guide sample: its type shows. Keyboard focus shows it too.
- [ ] `narra error` samples show the message inline, matching the `// error:` line.
- [ ] Build time and page weight before and after, in outcome.md.
