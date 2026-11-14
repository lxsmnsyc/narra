---
id: std-regex
title: "std/regex: syntax and a linear-time engine"
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
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent: 244852
owns:
  - std/src/regex.nra
  - selfhost/src/stdregex.nra
  - testdata/regex*.nra
  - testdata/regex*.out
  - testdata/errors/regex_*
  - roadmaps/selfhost/nodes/std-regex/**
shared:
  - selfhost/src/stdlib.nra
  - selfhost/src/stdtext.nra
  - testdata/errors/unknown_module.err
  - testdata/declsdump/imports/err_unknown_module.dump
  - SPEC.md
  - compiler/README.md
  - docs/guide/
reserves: [diag E1350-E1359]
provides: []
depends_on:
  - on: std-design
    kind: hard
    why: D118 item 6 fixes the syntax, the engine and the functions

---

# std/regex: syntax and a linear-time engine

## Goal

Build `std/regex` (alias `Regexes`, type `Regex`) as D118 item 6 says, without the compile-time check (node std-regex-const): a parser for JavaScript pattern syntax without back references and lookaround, a Pike VM in Narra with captures and leftmost-first matches, and `compile`, `test`, `find`, `findFrom`, `findAll`, `group`, `namedGroup`, `replace`, `replaceAll`, `replaceWith` and `split`.

## Context

- Source: D118 item 6. Russ Cox, "Regular Expression Matching: the Virtual Machine Approach", for the Pike VM with captures; RE2's syntax page for the inline flags and repeat limit.
- Write `compile` so the D105 evaluator can later run it: no arrow functions or function values, no traits beyond derived ones, only `std/string`, `std/array`, `std/map`, `std/int`, `std/range` calls where possible (std-regex-const adds `std/char`, `std/option` and `std/result` to the evaluator if the engine needs them). `Regex` holds only arrays, ints, strings and enums.

## Steps

- [ ] Pattern parser to an AST, then to a program of instructions (char, class, split, jump, save, assert, match). `RegexError { position, reason }` for every bad pattern; back references and lookaround give a reason that says the engine is linear-time; `{n,m}` above 1000 and programs above 10,000 instructions are errors.
- [ ] Pike VM over code points, with the literal-prefix search through `Str.indexOf`. `(?i)` folds ASCII only; `(?m)` and `(?s)` as in JavaScript's `m` and `s`.
- [ ] The functions of D118 item 6; `findAll` is a `Generator<Match>` that moves past an empty match by one character; `replaceWith` takes a non-escaping callback.
- [ ] Registry row, `stdlib.nra`, regenerated `stdtext.nra`, module-list goldens; SPEC, README and guide text.

## Acceptance

- [ ] `testdata/regex.nra`: 60 or more pattern and text pairs with their match and groups (literals, classes, anchors, `\b`, lazy and greedy repeats, alternation priority, named groups, empty matches, non-ASCII text).
- [ ] `testdata/regex_linear.nra`: `(a*)*b` against 10,000 `a` characters and `(a|aa)*c` against 5,000 finish (time printed only as pass or fail against a 1 s bound, so the output is stable).
- [ ] `testdata/regex_errors.nra`: each bad pattern prints its `RegexError` (unclosed group, bad class range, `\1`, `(?=`, `(?<=`, `{2000}`).
- [ ] `testdata/regex_replace.nra`: `replace`, `replaceAll`, `replaceWith` and `split`, including `$1` kept as text.
- [ ] Error test: `testdata/errors/regex_alias_clash.nra` (`import * as Regex from "std/regex"` beside `import { Regex }`, the existing alias error).
- [ ] Fast check passes: `sh scripts/test.sh`

## Sizing

L: about 1,500 lines of Narra; scores not low: complexity (engine), testing breadth.

## Notes for the worker

- Codes E1350-E1359 are reserved; none is expected (pattern errors are values).
- Offsets are byte offsets as `Str.indexOf` returns them.
- Do not import `std/unicode`.
