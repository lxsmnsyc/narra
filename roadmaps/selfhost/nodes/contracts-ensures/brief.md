---
id: contracts-ensures
title: "Contracts: ensures as an output pattern"
status: done
priority: should
type: feature
story: batch-7
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
spent: 660608
owns:
  - roadmaps/selfhost/nodes/contracts-ensures/**
shared:
  - selfhost/src/parser.nra
  - selfhost/src/ast.nra
  - selfhost/src/astdump.nra
  - selfhost/src/check.nra
  - selfhost/src/body.nra
  - selfhost/src/calls.nra
  - selfhost/src/stmts.nra
  - selfhost/src/operands.nra
  - selfhost/src/match.nra
  - selfhost/src/patterns.nra
  - selfhost/src/fmt.nra
  - selfhost/src/fmtexprs.nra
  - selfhost/src/diagnostics.nra
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0987, diag E0988]
provides: []
depends_on:
  - on: contracts
    kind: hard
    why: same parser, checker and prover code; builds on its requires facts
---

# Contracts: ensures as an output pattern

## Goal

D95, contracts part 2: `ensures <pattern>` with no named subject: `ensures >= 0`, `ensures Some(>= 0) | None`, `ensures Ok(> 0) | Err(_)`, `ensures #[>= 0, _]`, `ensures Rect { width: > 0 }`, `ensures "GET" | "POST"`. Every return must be proven to match. Comparison patterns need an ordered type; literal and constructor patterns work for any type with `==` and literals. Callers get the pattern as a fact about the result (so `if (r is Some(n))` after `ensures Some(>= 0) | None` knows `n >= 0`). On a generator it holds for each yielded value; on an async function it is not built yet (no async), so reject it there.

## Context

- Facts: the prover stores integer facts per local (`Fact` in `operands.nra`, see the contracts outcome). Proving an `ensures` means checking each `return` value against the pattern: comparison parts against the returned integer's fact or constant, constructor and literal parts structurally. Comparison patterns apply to integer positions only in this node, matching the prover; elsewhere they are an error at the declaration.
- Caller facts: a call's result is not a local, so the pattern becomes a fact only where the result is bound to an integer local: `const n = f()` (seed `n` from the comparison), and pattern bindings of the result (`if (f() is Some(n))`, `match` arms, `const r = f(); if (r is Some(n))` while `r` is not reassigned), which seed the binding local's fact from the matching alternative of the pattern. Bindings are locals, so the existing store holds them.

- Comparison patterns (`>= 0`) are new pattern syntax; allow them only in `ensures` in this node unless SPEC already allows them in `match`.
- Or-patterns exist (D98).

## Acceptance

- [ ] Samples for each D95 form, proven and unproven, and one caller using the fact.
- [ ] Fast check passes: `sh scripts/test.sh`
