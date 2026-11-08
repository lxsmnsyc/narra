---
id: checker-fixes
title: Fix suggestions in narrac diagnostics
status: done
priority:
type: feature
story: c-seed
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-08
spent: 291874
owns:
  - selfhost/src/fixes.nra
  - selfhost/src/check.nra
  - selfhost/src/parser.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/*.nra (fix sites only)
  - roadmaps/selfhost/nodes/checker-fixes/**
shared:
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: checker-codes
    kind: hard
    why: fixes attach to coded diagnostics and the JSON it adds
---

# Fix suggestions in narrac diagnostics

## Goal

Port the fix builders of `compiler/internal/narra/diagnostic_fixes.go` (importFix, concatFix, pushFix, quoteFix, lengthFix and the parser fixes, 44 sites) to the self-hosted compiler, so its diagnostics carry the same fix edits as Go's, in text and in `check --json`.

## Context

- Read `roadmaps/selfhost/nodes/seed-parity-spike/outcome.md` first: it has the measurements and the proposal this brief comes from.
- 65 samples carry a fix today. Go tests TestHabitFixes, TestHabitCases and TestCtorImportFix check them.

## Acceptance

- [ ] `narrac check --json` equals Go's on all 65 samples with a fix, and on every other error sample (no regression).
- [ ] An apply-fix path exists for tests (as Go has), and applying each fix gives the same file as Go.
- [ ] Fast check passes, as in checker-codes.

## Sizing

M: 44 fix sites with small builders.

Landed on feat/c-seed. Full runner passed after rebase onto goldens-tools.
