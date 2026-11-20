---
id: cli-reference
title: "Commands reference and complete usage text"
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
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/cli-reference/**
  - docs/reference/commands.md
shared:
  - selfhost/src/driver/driver.nra
  - selfhost/src/main.nra
  - selfhost/tests/src/**
  - docs/guide/18-tools.md
reserves: []
provides: []
depends_on:
  - on: site-reference-section
    kind: hard
    why: the page lives in its section
  - on: guide-targets
    kind: hard
    why: both edit chapter 18; land after it
---
# Commands reference and complete usage text

## Goal

One complete usage text in `narrac` covering `fmt`, `query`, `lsp`, `apply-fix`, `emit`; bare `narrac` prints it. `/reference/commands` by hand, with a test that every command and flag in `narrac --help` appears. Chapter 12's table becomes a link. Proposal change 13.

Outline from the approved docs-structure-review proposal (`roadmaps/selfhost/nodes/docs-structure-review/proposal.md`). Refined 2026-11-20 into batch-21 (docs-concept-audit, approved).

## Acceptance

- [ ] `narrac --help` lists every command; `/reference/commands` documents every command and flag with exit codes and environment variables; a test checks the page covers the usage text; chapter 18's table links to it.
- [ ] `sh scripts/test.sh` and `pnpm --dir website build` pass.
