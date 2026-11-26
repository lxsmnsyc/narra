---
id: docs-packages
title: "docs: a packages section on the site, and [c] in the narra.toml reference"
status: active
priority: could
type: docs
story: batch-26
early: false
outline: false
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: S
risk: low
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent:
owns:
  - roadmaps/selfhost/nodes/docs-packages/**
  - website/**
  - docs/packages/**
shared:
  - docs/reference/narra-toml.md
  - docs/guide/08-modules-and-packages.md
  - docs/README.md
  - packages/*/README.md
  - selfhost/tests/src/*.nra
reserves: []
provides: []
depends_on:
  - on: c-packages
    kind: hard
    why: documents the [c] table and C packages as built
---

# docs: a packages section on the site, and [c] in the narra.toml reference

## Goal

The site shows the first-party packages (http, websocket, yaml, changesets, and any later `packages/*`), and the `narra.toml` reference covers the `[c]` table.

## Context

- Maintainer, 2026-11-26: the packages have READMEs (`packages/<name>/README.md`) that the site never shows. `website/src/pages.ts` globs only `docs/guide/[0-9]*.md` and `docs/reference/**/*.md`; search is in `website/plugins/search.ts`; `website/scripts/vercel-ignore.sh` decides which paths trigger a deploy.
- Pages come straight from `packages/*/README.md`, so there is one copy of each text. Add a packages index (for example `docs/packages/README.md`, or generated from the READMEs' first paragraph) and a route per package, in the site navigation and search. A new `packages/<name>/README.md` must appear with no site edit.
- `vercel-ignore.sh` must deploy when a package README changes.
- `docs/reference/narra-toml.md` lacks `[c]` (`sources`, `include`, `defines`), `[c.os.<Os>]`, `[c] git` and `tag` for C packages, and the `extern {} from` forms (`"./file.c"` relative to the declaring module, `"<package>/<path>"`, a library name). Take the facts from SPEC sections 9 and 15 and the outcomes of `package-c-sources` and `c-packages`; write them as reference text, standalone (no SPEC or D-number citations, no assumed TypeScript knowledge).
- Guide chapter 8 gets one short paragraph pointing to the packages section.
- Links inside a package README that point at repo files (`../../examples/http_ping`) must work on the site or be rewritten to a working link at build time.
- If a docs suite checks links or page structure (`selfhost/tests` or a website check), extend it to the new pages.

## Acceptance

- [ ] `pnpm build` in `website/` passes, and the built site has a packages index and one page per `packages/*/README.md`, reachable from the navigation and found by search.
- [ ] No link on the new pages is broken (checked by the site build or a link check).
- [ ] `vercel-ignore.sh` deploys on a package README change.
- [ ] `docs/reference/narra-toml.md` documents `[c]`, `[c.os.<Os>]`, `[c] git`/`tag` and the `from` forms.
- [ ] `sh scripts/test.sh` passes.
