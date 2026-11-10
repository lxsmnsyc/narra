---
id: parser-token-kinds
title: "The parser compares token kinds, not text"
status: active
priority: should
type: feature
story: batch-5
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-10
spent:
owns:
  - roadmaps/selfhost/nodes/parser-token-kinds/**
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
reserves: []
provides: []
depends_on:
  - on: lexer-punctuation
    kind: hard
    why: both edit how the lexer classifies punctuation
---

# The parser compares token kinds, not text

## Goal

`isAt` and `parseBinary` in `selfhost/src/parser.nra` compare token text with `memcmp`. Give punctuation and keyword tokens an int kind at lex time and compare that.

## Context

- Evidence: `isAt` 2.4% inclusive; `memcmp` from `isAt` 143 and `parseBinary` 67 samples. Saves about 1.5%. Read `git show feat/batch-5:roadmaps/selfhost/nodes/profile-compiler-2/outcome.md`, proposal 9.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, run 3 times each side, reported in the outcome. The gain must exceed the run-to-run spread, or the node closes as not worth it.
- [ ] Parse results are unchanged: emitted C for `selfhost` byte-identical, fmt and error suites pass.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Other batch-5 nodes change the compiler at the same time, so measure the base again at your own starting commit.
- Do not regenerate the seed.
