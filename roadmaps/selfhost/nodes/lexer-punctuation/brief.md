---
id: lexer-punctuation
title: "The lexer picks punctuation by first byte"
status: todo
priority: should
type: feature
story: batch-5
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
  - roadmaps/selfhost/nodes/lexer-punctuation/**
shared:
  - selfhost/src/lexer.nra
  - testdata/lexer_*
reserves: []
provides: []
depends_on:
  - on: profile-compiler-2
    kind: hard
    why: the profile is the evidence for this change
---

# The lexer picks punctuation by first byte

## Goal

`lexPunctuation` (`selfhost/src/lexer.nra:453`) picks candidates from a table indexed by the first byte, longest first, instead of trying every entry of `PUNCTS` with `hasPrefix`.

## Context

- Evidence: `lexPunctuation` 5.7% and `hasPrefix` 5.3% inclusive at -fno-inline; the lex phase is 7.2%. Saves 3% to 4%. Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 3.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. The gain must exceed the run-to-run spread, or the node closes as not worth it.
- [ ] Token streams are unchanged: `narrac` output for the samples and the fmt suite is byte-identical.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Other batch-5 nodes change the compiler at the same time, so measure the base again at your own starting commit.
- Do not regenerate the seed.
