---
id: sized-conversions
title: "Integer conversions: I32.from, tryFrom and wrap"
status: done
priority: should
type: feature
story: batch-1
early: false
outline: false
milestone: m12-spec-gaps
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
added: 2026-11-07
spent: 118762
owns:
  - selfhost/src/stdsized.nra
  - testdata/sized_convert*
  - testdata/errors/sized_convert_*
  - roadmaps/selfhost/nodes/sized-conversions/**
shared:
  - selfhost/src/stdint.nra
  - selfhost/src/stdlib.nra
  - SPEC.md
  - compiler/README.md
reserves: [diag E0940, diag E0941]
provides: []
depends_on:
  - on: sized-numbers
    kind: hard
    why: needs the integer types in both checkers and emitters; its gate fixes the names (D82)
  - on: std-tables-split
    kind: hard
    why: adds seven std modules, each a file of its own after the split
---

# Integer conversions: I32.from, tryFrom and wrap

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

The conversion functions of D82 in both compilers: modules `std/i8`, `std/i16`, `std/i32`, `std/u8`, `std/u16`, `std/u32`, `std/u64` (imported as `I8` ... `U64`), each with `from`, `tryFrom`, `wrap`, `toString` and `parse`, and `from`, `tryFrom` and `wrap` added to `std/int`.

## Context

- Read `nodes/sized-numbers/outcome.md` (and D82) and `nodes/std-tables-split/outcome.md` first.
- These are builtins (C casts and range checks), registered in both compilers like `std/int`; `TestSelfhostBuiltins` compares the registries. One Go file (`std_sized.go`) can register all seven modules from a table.
- `Int.parse` and `Int.toString` (`std_int.go`) are the pattern for `parse` and `toString`; `parse` errors keep the `std/int` wording with the type's range.

## Acceptance

- [ ] `X.from(v)` accepts an argument of any integer type whose every value fits `X`; any other integer type is E0940 `I32.from cannot take u32 without loss; use I32.tryFrom or I32.wrap`. A non-integer argument is E0941.
- [ ] `X.tryFrom(v): Option<X>` and `X.wrap(v): X` accept any integer type. `wrap` keeps the low bits, as a C cast.
- [ ] Sample `sized_convert.nra` and `.out`: every lossless pair `from` accepts (a table driven by a loop is fine), `tryFrom` at each boundary (`MAX`, `MAX + 1`, `MIN`, `MIN - 1` through a wider type), `wrap` of negative values into unsigned types and back, `parse` and `toString` round trips at the limits. ASan and leak check clean.
- [ ] Error tests for E0940 and E0941.
- [ ] SPEC "Standard modules" lists the modules; README "Supported" describes them. Go and `selfhost/` emit the same C; fast check passes.

## Sizing

M. Seven small builtin modules from one table in each compiler; uncertainty low, tests automatic.

Landed on feat/batch-1. Gate choice to review: a literal argument to `from` is typed `int`, so `I32.from(5)` is E0940.
