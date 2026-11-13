---
id: std-unicode-normalize
title: "std/unicode: normalization"
status: todo
priority: should
type: feature
story: batch-14
early: false
outline: false
milestone: p2-std
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
added: 2026-11-13
spent:
owns:
  - testdata/unicode_normalize*.nra
  - testdata/unicode_normalize*.out
  - testdata/unicode/NormalizationTest-sample.txt
  - roadmaps/selfhost/nodes/std-unicode-normalize/**
shared:
  - std/src/unicode.nra
  - std/src/unicode__tables.nra
  - scripts/unicode/generate.nra
  - selfhost/src/stdtext.nra
  - SPEC.md
  - docs/guide/
reserves: [diag E1325-E1329]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 3 fixes the normalization API
  - on: std-unicode
    kind: hard
    why: extends its generator, table format and module
---

# std/unicode: normalization

## Goal

Add `normalize(text, form: NormalForm): string` and `isNormalized(text, form): bool` with `enum NormalForm { NFC, NFD, NFKC, NFKD }` to `std/unicode` (D118 item 3), with the decomposition, composition and combining-class tables from the std-unicode generator. Split from std-unicode because these tables are the largest (about 70 KB) and the least needed by tools; per D118 question (f) this node lands last in the batch and may move out if the batch runs long.

## Context

- Source: D118 item 3; UAX #15.
- Counts at Unicode 13: 2,061 canonical and 3,675 compatibility decompositions, 549 combining-class ranges (std-design log).

## Steps

- [ ] Extend `scripts/unicode/generate.nra` with `UnicodeData.txt` decompositions and combining classes, `CompositionExclusions.txt` and the quick-check properties (`DerivedNormalizationProps.txt`); Hangul is computed, not tabled.
- [ ] `normalize` and `isNormalized` (quick check first, then the full test), ASCII fast path that returns the text unchanged with no allocation.
- [ ] Commit a sample of `NormalizationTest.txt` (part 0, part 1 up to 2,000 lines) as `testdata/unicode/NormalizationTest-sample.txt`.
- [ ] SPEC and guide text as built; record the size growth of narrac and of a program that calls `normalize`.

## Acceptance

- [ ] `testdata/unicode_normalize.nra` checks the five columns of every sample line in all four forms (none fails).
- [ ] `testdata/unicode_normalize_basic.nra`: `normalize("e\u{301}", NormalForm.NFC)` is `é`, NFKC of `ﬁ` is `fi`, Hangul syllables round-trip, ASCII unchanged.
- [ ] A program that uses only `graphemes` keeps the same binary size as before this node (tree shaking leaves the normalization tables out).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: one algorithm and its tables over an existing generator.

## Notes for the worker

- Codes E1325-E1329 are reserved; none is expected.
