---
id: template-one-alloc
title: "Build template literals in one allocation"
status: todo
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
  - roadmaps/selfhost/nodes/template-one-alloc/**
shared:
  - selfhost/runtime/runtime.h
  - selfhost/src/runtimecore.nra
  - selfhost/src/emitexprs.nra
  - testdata/template_one_alloc*
reserves: []
provides: []
depends_on:
  - on: perceus
    kind: hard
    why: both edit emitexprs.nra; the earlier node lands first
  - on: int-to-string
    kind: hard
    why: both edit runtime.h; the earlier node lands first
  - on: profile-compiler
    kind: hard
    why: the profile is the evidence for this change
---

# Build template literals in one allocation

## Goal

Emit a template literal with n parts as one runtime call that sums the lengths and copies each part once (for example `lstr_concat_n(count, parts)`), instead of n-1 `lstr_concat` calls that allocate n-1 times and copy the prefix again at each step. String `+` chains can use the same call if it is simple.

## Context

- Evidence: `lstr_concat` 11.3% inclusive (`memmove` 3.6%, `malloc` 6.0%), plus freeing the intermediate strings, spread over `positionKey`, `writeLine`, `internedName`, `localName` and `temp`. Profile proposal 5, saves about 5% after int-to-string, effect-int-keys and cache-type-names. Measured at ffdea9c (feat/batch-4); read `git show feat/batch-4:roadmaps/selfhost/nodes/profile-compiler/outcome.md`.
- Other batch-4 nodes change the compiler at the same time, so measure the base again at your own starting commit.

## Acceptance

- [ ] Instructions retired for `narrac emit selfhost` (macOS `/usr/bin/time -l`) before and after on the same commit, reported in the outcome. The gain must exceed the run-to-run spread (run each side 3 times).
- [ ] A sample with templates of 1, 2 and many parts, empty parts, non-ASCII text and interpolated ints, clean under AddressSanitizer and the leak check.
- [ ] `sh scripts/test.sh` passes, clean under AddressSanitizer and the leak check.

## Sizing

M: runtime or lexer core with edge cases (see acceptance).

## Notes for the worker

- Behavior-preserving: no language change, no new diagnostics.
- Do not regenerate the seed.
