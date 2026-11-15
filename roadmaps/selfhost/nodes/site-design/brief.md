---
id: site-design
title: "Site design system, motion and home page"
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
  - roadmaps/selfhost/nodes/site-design/**
  - website/src/styles.css
  - website/src/routes/index.brm
  - website/src/routes/_layout.brm
  - website/src/routes/_error.brm
  - website/src/components/Icon.brm
  - website/src/home/**
  - website/src/motion/**
  - website/public/**
shared:
  - website/src/theme.ts
reserves: []
provides: []
depends_on:
  - on: curl-installer
    kind: contract
    via: install-command
---
# Site design system, motion and home page

## Goal

A visual identity and motion language for the site, then the home page on it. Tokens (color, type scale, spacing in steps of 4/8, radius, shadow, motion durations and easings), header, footer, buttons, code panels. Motion responds to the user and to scroll position, uses only transform and opacity, reverses mid-way, and honors prefers-reduced-motion (ui-style skill).

Home page content is what a learner values: what the code looks like, what the compiler catches for them (errors that say what to write, exhaustive matches, no null, mutable value checks), the tools they get (fmt, test, lsp, fix). Not how narrac is built. The hero shows the install one-liner from contract install-command with a copy button and an OS switch.

Every snippet and compiler message on the page is run through narrac.

## Acceptance

- [ ] Tokens and motion primitives in styles.css, used by the home page and the layout.
- [ ] Home: hero with the one-liner, at least three learner-facing features with real narrac output.
- [ ] Checked in Chrome: dark, light, phone width, reduced motion. No layout shift on load.
