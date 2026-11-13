---
id: std-unicode
title: "std/unicode: tables, graphemes, case and width"
status: active
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
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - std/src/unicode.nra
  - std/src/unicode__tables.nra
  - scripts/unicode/**
  - selfhost/src/stdunicode.nra
  - testdata/unicode/**
  - testdata/unicode*.nra
  - testdata/unicode*.out
  - testdata/errors/unicode_*
  - roadmaps/selfhost/nodes/std-unicode/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/stdtext.nra
  - testdata/errors/unknown_module.err
  - testdata/declsdump/imports/err_unknown_module.dump
  - SPEC.md
  - compiler/README.md
  - docs/guide/
reserves: [diag E1320-E1324]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 3 fixes the functions, the Unicode version and the table form
  - on: std-foundation
    kind: hard
    why: source std modules need the generic mismatch errors, std/internal and the stage 2 tests (std-source gate, 2026-11-05)
  - on: std-source
    kind: hard
    why: the module is Narra source
---

# std/unicode: tables, graphemes, case and width

## Goal

Build `std/unicode` (alias `Unicode`, Unicode 17.0) as D118 item 3 says, without normalization (node std-unicode-normalize): the table generator, `graphemes`, `graphemeCount`, `toUpperCase`, `toLowerCase`, `foldCase` and `displayWidth`. Tables are string constants of fixed-width ASCII records in generated Narra source.

## Context

- Source: D118 item 3; SPEC section 3 "Std modules for tools" and section 4 "Strings".
- Evidence: a 6,000-element `Array<int>` literal constant takes 52 s to build (clang on the D61 getter); a 120,000-character string constant takes 0.15 s (`roadmaps/selfhost/nodes/std-design/log.md`). Do not write tables as array literals.

## Steps

- [ ] `scripts/unicode/generate.nra`: reads `UnicodeData.txt`, `SpecialCasing.txt`, `CaseFolding.txt`, `GraphemeBreakProperty.txt`, `emoji-data.txt`, `EastAsianWidth.txt` and `DerivedCoreProperties.txt` (Indic conjunct break, UAX #29 rule GB9c) from a folder given as its argument, and writes `std/src/unicode__tables.nra`. The header names the version and the SHA-256 of each input. Records are ranges, fixed width, sorted, so a lookup is a binary search with `Str.byteAt`. Leave room for the normalization tables of the next node.
- [ ] Try `std/internal/unicode` as the module path of the tables (a source module under `internal`, so programs cannot import it). If the loader allows only builtin internal modules, use `std/unicode/tables` and say so in the outcome.
- [ ] `std/src/unicode.nra`: `graphemes(text): Generator<string>` (slices of `text`), `graphemeCount`, `toUpperCase`, `toLowerCase` (full mappings, unconditional SpecialCasing rules, final sigma, no locale), `foldCase` (full folding, status C and F), `displayWidth` (per cluster: Wide and Fullwidth 2, marks, format characters and zero width 0, emoji presentation 2, others 1). ASCII fast paths in every function.
- [ ] Commit `testdata/unicode/GraphemeBreakTest.txt` with the Unicode License v3 notice beside it.
- [ ] Registry row, `stdlib.nra`, regenerated `stdtext.nra`, module-list goldens. Record narrac's size growth from the embedded tables in the outcome.
- [ ] SPEC section 3 and section 4 "Strings" text as built; README; the guide's standard library chapter.

## Acceptance

- [ ] `testdata/unicode_graphemes.nra` runs every line of `GraphemeBreakTest.txt` and prints the count of passes and the failing lines (none).
- [ ] `testdata/unicode_case.nra`: `toUpperCase("straße")` is `STRASSE`, `toLowerCase("ΣΑΣ")` ends with a final sigma, `foldCase` makes `Straße` and `STRASSE` equal, ASCII and empty strings.
- [ ] `testdata/unicode_width.nra`: widths of ASCII, CJK, a combining mark, a flag and a ZWJ family emoji.
- [ ] Error test: `testdata/errors/unicode_internal_import.nra` (a program imports the tables module: E0426, or the error of the chosen path).
- [ ] Running the generator twice on the same inputs gives the same file (a step in the sample or a test in the `generated` suite).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

L: a generator program, seven input formats, four algorithms and a conformance test. Scores not low: breadth (new script plus module), unknowns (table encoding).

## Notes for the worker

- Download the UCD 17.0.0 files once into the scratch folder; never commit them except the test file above.
- Codes E1320-E1324 are this node's; E1325-E1329 are std-unicode-normalize's. None is expected.
- Do not import `std/unicode` from `std/terminal` or `std/regex` (D118).
