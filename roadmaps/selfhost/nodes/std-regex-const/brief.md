---
id: std-regex-const
title: "std/regex: patterns checked while compiling"
status: done
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
spent: 177421
owns:
  - testdata/regex_const*.nra
  - testdata/regex_const*.out
  - testdata/errors/regex_checked_*
  - roadmaps/selfhost/nodes/std-regex-const/**
shared:
  - std/src/regex.nra
  - selfhost/src/stdregex.nra
  - selfhost/src/consteval.nra
  - selfhost/src/constfunctions.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E1360-E1369]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 6 fixes Regexes.checked and E1360
  - on: std-regex
    kind: hard
    why: checked runs the engine's compile in the D105 evaluator
---

# std/regex: patterns checked while compiling

## Goal

Add `Regexes.checked(pattern): Regex`, an `@const` function: as the value of a top-level `const` the compiler runs it (D105), and a bad pattern stops the evaluation with E1053 and the `RegexError` text. Anywhere else it is E1360, so it never panics at run time (D92).

## Context

- Source: D118 item 6; D105 items 3, 6 and 7 (allowed modules, panics as compile errors, the C writer).
- The evaluator supports `std/string`, `std/array`, `std/map`, `std/int`, `std/float` and `std/range` (node const-functions). D105 item 3 lists `std/char`, `std/option` and `std/result` as allowed but not built.

## Steps

- [ ] Add to the evaluator the rows of `std/char`, `std/option` and `std/result` that `std/regex`'s compile path calls (no function-taking ones), and allow `std/regex` source bodies.
- [ ] `checked` in `std/src/regex.nra`, marked `@const`, panicking with the pattern error.
- [ ] E1360 in `diagnostics.nra` (`Regexes.checked can only be the value of a top-level const; use Regexes.compile`), with an explain text, reported by a check on the registry row of `checked`.
- [ ] SPEC text as built.

## Acceptance

- [ ] `testdata/regex_const.nra`: `const WORD = Regexes.checked("[a-z]+")` matches at run time, and the same pattern from `compile` at run time gives the same matches (the D105 run-twice rule).
- [ ] Error tests: `testdata/errors/regex_checked_bad.nra` (`checked("(a")` in a const: E1053 with the position and reason), `testdata/errors/regex_checked_runtime.nra` (`checked` inside a function: E1360).
- [ ] The step count of a 40-character pattern stays under 1% of the D105 budget (printed in the outcome).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

M: evaluator rows for three modules, one check and one code.

## Notes for the worker

- `consteval.nra` is a hot file for const work; rebase before review.
- Codes E1360-E1369: E1360 is used; the rest stay free.
