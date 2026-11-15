---
id: drop-npm-narrac
title: "Remove the npm packages of narrac"
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
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent:
owns:
  - roadmaps/selfhost/nodes/drop-npm-narrac/**
  - npm/**
  - scripts/npm-package.sh
  - README.md
  - DECISIONS.md
shared:
  - roadmaps/selfhost/GOAL.md
reserves: []
provides: []
depends_on: []
---
# Remove the npm packages of narrac

## Goal

Maintainer, 2026-11-16: the npm packages of narrac misread the goal. Go and Rust do not ship their compilers through npm. esbuild and oxc ship npm packages because those projects chose to, with their own scripts. narrac ships as release archives and the curl installer (curl-installer). Remove `npm/` and `scripts/npm-package.sh`, and revise D123 item 8 to answer how narrac and Narra packages are distributed.

## Acceptance

- [x] `npm/` and `scripts/npm-package.sh` removed; README and GOAL no longer point at them.
- [x] D123 item 8 says how narrac, Narra libraries, Narra programs and JS-facing builds are distributed.
