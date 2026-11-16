---
id: vercel-hovers
title: "Hovers in the Vercel build"
status: done
priority: could
type: chore
story: batch-17
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
verify: sh website/scripts/vercel-build.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/vercel-hovers/**
  - website/scripts/vercel-build.sh
  - website/vercel.json
shared: []
reserves: []
provides: []
depends_on:
  - on: narra-twoslash
    kind: hard
    why: builds the narrac its plugin calls
---
# Hovers in the Vercel build

## Goal

The Vercel build image (Amazon Linux 2023) has no C compiler, so a deployed site would lose the twoslash hovers. `website/scripts/vercel-build.sh` installs gcc with dnf when `cc` is missing, compiles `seed/narrac.c` once (the seed is current, so no bootstrap stages), sets `NARRAC`, and builds. Without a compiler it still builds, with plain samples. `vercel.json` runs it, and its ignore command now also watches `seed/` and `std/src`, since hovers change with them.

## Acceptance

- [x] Locally: 35 s to compile the seed, then the build has hovers (27 on the home page).
- [ ] On Vercel: not run; agents do not deploy. The first deploy after merge shows whether dnf works in the build step.
