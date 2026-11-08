---
id: sized-numbers
title: "Sized integer types i8 to u64"
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-04
spent:
owns:
  - testdata/sized_int*
  - testdata/errors/sized_int_*
  - roadmaps/selfhost/nodes/sized-numbers/**
shared:
  - selfhost/runtime/**
  - selfhost/src/types.nra
  - selfhost/src/resolve.nra
  - selfhost/src/literals.nra
  - selfhost/src/exprs.nra
  - selfhost/src/traits.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emittraits.nra
  - SPEC.md
  - compiler/README.md
  - DECISIONS.md
reserves: [diag E0294, diag E0295, diag E0296, decision D82]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point (done)
---

# Sized integer types i8 to u64

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

Spec-gap F7, integer part (SPEC section 4 "Primitive types" and "Numbers"): `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32` and `u64` are unknown types today, and "a literal must fit its type" holds only for `int`. Add the types, literal typing and arithmetic in both compilers, and record the numeric design, conversions included, as D82. Conversions are node sized-conversions, `f32` is sized-f32, `bytes` is bytes-type.

## Context

- Spec-gap audit row F7: `nodes/spec-gap-audit/outcome.md`. `int` is `PrimType{"int"}` (`types.go`); the literal check is E0200 in `check_expr.go`; checked arithmetic is `__builtin_*_overflow` in `runtime.h`, which works for any C integer type.
- Selfhost: the `Type` enum in `selfhost/src/types.nra`, literal checks in `literals.nra` and `exprs.nra`.

## Rules

1. `i64` is another name for `int`: the same type, printed `int`. The others map to `int8_t` ... `uint64_t`.
2. An integer literal takes the integer type the context expects (annotation, parameter, field, element, other operand); with none it is `int`. It must fit that type: `const x: u8 = 256` and `const y: u8 = -1` are E0294 `integer literal 256 does not fit in u8`.
3. No implicit conversion. Arithmetic, comparison and bitwise operands have the same type (E0295 names both types and suggests the conversion functions of sized-conversions). Unary `-` on an unsigned type is E0296.
4. Overflow panics in debug builds and wraps in release builds, at the type's width; `+%`, `-%` and `*%` wrap at the width. Division and remainder by zero panic, and so does `MIN / -1` in debug builds.
5. Every integer type implements `Equatable`, `Hashable`, `Ordered` and `Show`, works as a map key, a struct field, a tuple element, an enum field and a generic argument, and prints in template strings. `Array<u8>` stores one byte per element.

## Acceptance

- [ ] Sample `sized_int.nra` and `.out`: each type's minimum and maximum written as literals and printed (`u64` above `int`'s maximum included); arithmetic in each type; wrapping operators at the width; a struct and an `Array<u8>` of sized fields; a `Map<u16, string>`; `==`, ordering and `Show` of a derived struct with sized fields. Clean under ASan and the leak check. A test (`std/test` panics form) shows `i8` overflow panics in a debug build.
- [ ] Error tests, one per code, plus an `i32` passed to an `int` parameter (E0119).
- [ ] DECISIONS.md D82 records the decision below; SPEC section 4 "Numbers" states rules 2 to 4 and replaces the placeholder `Int.toI32(x)` with the names of D82.
- [ ] Go and `selfhost/` emit the same C; fast check passes including `TestSelfhostBootstrap`.

## Sizing

M. Uncertainty low (Go and Rust rules), blast radius medium (the primitive type set in both checkers and emitters), size medium, tests automatic.

## Notes for the worker

- Context typing of literals reaches every place `want` is passed: binary operands (type the literal from the other side), returns, struct and array literals, default field values, `match` patterns on integers.
- Printing a `u64` needs `PRIu64`; `Show` for every width goes through one runtime helper per signedness.

## Decision to confirm

Plan (D82): Go and Rust's model. Context-typed literals, no implicit conversion, overflow as for `int`. Conversions are functions named after the target type, one module per sized type, built by sized-conversions:

| Function | Meaning |
| --- | --- |
| `I32.from(x)` | from any integer type, valid only when every value fits (`i8`, `i16`, `u8`, `u16` to `i32`); anything else is a compile error that names the next two |
| `I32.tryFrom(x)` | from any integer type, `Option<i32>`, `None` when out of range |
| `I32.wrap(x)` | from any integer type, keeps the low bits (two's complement), as a C or Go cast |
| `I32.toString(x)`, `I32.parse(text)` | as `Int.toString` and `Int.parse` |

Modules `std/i8`, `std/i16`, `std/i32`, `std/u8`, `std/u16`, `std/u32`, `std/u64`, imported as `I8` ... `U64`. `std/int` serves `i64` and gets `from`, `tryFrom` and `wrap`. `std/f32` (`F32.from`, `F32.truncate`) and `Float.from` cover floats (sized-f32).

- Reason: the lossless and lossy cases have different names, so a truncating conversion is visible at the call. Rust has the same split (`From`, `TryFrom`, `as`).
- Reason: one module per type follows `Int` and `Float`, and SPEC's original `I32.from(x)`.
- Rejected: implicit widening, as in C. Mixed-type arithmetic hides where values change size.
- Rejected: `as` casts. One form for lossless and lossy conversions, and a new keyword.
- Rejected: `Int.toI32(x)` style names in `std/int` (the spec placeholder). One function per pair of types, more than 60 names.
- Rejected: literal suffixes (`200u8`). Context typing covers them; they can be added later without breaking programs.
