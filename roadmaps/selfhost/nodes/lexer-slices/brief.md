---
id: lexer-slices
title: "Lexer reads bytes and slices token text"
status: active
priority: should
type: feature
story: batch-4
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
  - roadmaps/selfhost/nodes/lexer-slices/**
shared:
  - selfhost/src/lexer.nra
  - testdata/lexer_*
reserves: []
provides: []
depends_on:
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# Lexer reads bytes and slices token text

## Goal

The lexer works on the source string's bytes (`Str.byteAt`) and slices token text out of the source. Today `createLexer` decodes the whole file with `Str.chars`, `readText` builds each token's text one character at a time, and `punctuationChars` is rebuilt for every lexer, template parts included.

## Context

- Evidence: The lex phase is 12.2% of `narrac emit selfhost`: `readText` 4.9%, `createLexer` 2.4%, `runLexer` self 2.4%. Profile proposal 7, saves 6% to 8%. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] Line and column positions stay right for non-ASCII source: a sample or error test with multi-byte characters before an error shows the same position as on the base. The LSP suite (UTF-16 positions) passes.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

M: runtime or lexer core with edge cases (see acceptance).

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
