---
id: finish-facts-lines
title: "finishFacts without a full split and a second lex"
status: active
priority: should
type: fix
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/finish-facts-lines/**
shared:
  - selfhost/src/parser.nra
reserves: []
provides: []
depends_on:
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# finishFacts without a full split and a second lex

## Goal

`findHeaderLine` (`selfhost/src/parser.nra:261`) scans for the header instead of splitting the whole source into lines, and `collectReferencedNames` reuses the tokens it already has instead of calling `Lexer.lex` on template parts again.

## Context

- Evidence: `finishFacts` 4.9% inclusive: `narra_str_split` 2.1%, `collectReferencedNames` 1.8% (0.9% in `lexer.lex`). Profile proposal 8, saves about 3.5%. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] The parse, AST dump, unused-import and query suites are unchanged.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

S: one file, a measured hotspot.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
