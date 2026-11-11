---
id: const-generics
title: "Constant type parameters: const N: int"
status: done
priority: should
type: feature
story: batch-8
early: false
outline: false
milestone:
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
added: 2026-11-11
spent: 459155
owns:
  - roadmaps/selfhost/nodes/const-generics/**
  - selfhost/src/constgenerics.nra
shared:
  - selfhost/src/lexer.nra
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/check.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/idcheck.nra
  - selfhost/src/trivia.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - selfhost/tests/spec-skip.txt
  - selfhost/src/generics.nra
  - selfhost/src/instances.nra
  - selfhost/src/calls.nra
  - selfhost/src/resolve.nra
  - selfhost/src/declare.nra
  - selfhost/src/types.nra
  - selfhost/src/cnames.nra
  - selfhost/src/operands.nra
  - selfhost/src/traits.nra
  - selfhost/src/stdtypes.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/sharedresults.nra
  - selfhost/src/exprs.nra
reserves: [diag E0992, diag E0993]
provides: []
depends_on: []
---

# Constant type parameters: const N: int

## Goal

D95 fixed arrays, part 1: constant type parameters `const N: int` on functions and structs. `N` is inferred from argument or expected types, or written as `f<3>(x)`; `N` is also a constant value in the body, so the prover (`operands.nra`) treats it as a known constant per instance. No arithmetic on `N` in types. Each distinct value is its own generic instance.

## Context

- Design: DECISIONS.md D95 ("Fixed arrays"); SPEC section 3 (`function sum<const N: int>(values: FixedArray<f64, N>): f64`).
- Generic instances and their C names: `generics.nra`, `instances.nra`, `cnames.nra`. A const argument must be part of the instance key and of the hashed C name.
- Without FixedArray (next node) the feature is usable with a const parameter in plain functions and structs (`function repeat<const N: int>(text: string): string`), which is what the samples test.
- Integer kinds: `const N: int` only in this node; say in SPEC whether other integer kinds come later.

## Acceptance

- [ ] Samples: inference from an expected type, explicit `f<3>(x)`, `N` used as a divisor without a check (proven non-zero when the instance's value is not 0), two instances with different values in one program, a struct with a const parameter, and errors for arithmetic on `N` in a type, a non-constant argument, and a mismatch.
- [ ] Fast check passes: `sh scripts/test.sh`
