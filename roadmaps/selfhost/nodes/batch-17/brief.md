---
id: batch-17
title: "Batch 17: website, docs and the installer"
status: active
priority: could
type: story
story: 
early: false
outline: false
milestone: p6-docs
target:
ship: pr
land: auto
from:
size: L
spike: false
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/batch-17/**
shared:

reserves: []
provides: []
depends_on:
  - on: batch-16
    kind: hard
    why: the site, install.sh and release.yml landed there
---
# Batch 17: website, docs and the installer

## Goal

Maintainer request, 2026-11-16: the site is "still pretty boring". Redesign it as an award-level motion and UI designer would. Fix the docs pages and their inconsistent highlighting. The home page shows only what helps a learner (not how the compiler is built). Install is one `curl | sh` line, as on pnpm.io and gojo.egoist.dev, not "clone and bootstrap". Port twoslash (https://shikijs.github.io/twoslash/) to Narra: hover types and inline errors in code samples, computed at build time.

## Children

- site-design: design system, motion, header and footer, home page.
- docs-redesign: the guide, install and spec pages on the new design.
- docs-highlighting: every fence highlighted the same way, checked in the build.
- query-hovers: one narrac query that returns every hover type and diagnostic of a file.
- narra-twoslash: hover popups and inline errors on the site, from query-hovers.
- curl-installer: release assets install.sh and install.ps1, every prebuilt target, docs/install.md led by the one-liner.
- docs-search: a search palette over the guide, install and spec, as on the maintainer's other sites.
- drop-npm-narrac: remove narrac's npm packages.
- first-release (gate): push the first `v*` tag, so the one-liner works. The maintainer decides.

## Acceptance

- [ ] Every child done or out of the batch with a reason.
- [ ] `pnpm --dir website build` and `sh scripts/test.sh` pass; seed current if selfhost changed.
- [ ] Checked in Chrome: home, a guide chapter, install and spec, dark and light, phone width, reduced motion.
