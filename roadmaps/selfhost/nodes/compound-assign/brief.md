---
id: compound-assign
title: "Compound assignment, increment and decrement"
status: done
priority: should
type: feature
story: batch-2
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
added: 2026-11-02
spent: 194545
owns:
  - testdata/compound_assign*
  - testdata/errors/compound_assign*
  - roadmaps/selfhost/nodes/compound-assign/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [code E0962, code E0963, code E0964]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes go into the Narra compiler only, after the bootstrap fixed point
---

# Compound assignment, increment and decrement

## Goal

D76 (maintainer, 2026-11-04) keeps compound assignment. Build `x += e`, `-=`, `*=`, `/=`, `%=` and the bitwise and shift forms (`&=`, `|=`, `^=`, `<<=`, `>>=`), and `x++` / `x--`, on every target `=` accepts: a `let` binding, a field or index of a changeable value (D81), and a `mutable` parameter's contents.

Rules (write them into D76 as an amendment):

- Same meaning as `x = x op e`, with the target place evaluated once: `items[next()] += 1` and `items[next()]++` call `next` once.
- Statements only, like `=`. They have no value. Only the postfix spelling `x++` / `x--`; `++x` is an error with a hint.
- `++` and `--` take an integer target of any sized integer type. Overflow, division by zero and shift checks are the same as for the binary operator.
- String `+=` concatenates, and drops the old value's reference correctly.
- A Map index target (`counts[key] += 1`) with a missing key panics, as reading `counts[key]` does. Say that in SPEC.md.

New AST node kinds need checker, emitter, `fmt` and `astdump` cases.

Seed rule: do not use the new syntax in `selfhost/src` in this batch. The seed is regenerated at the end of the batch; using it in the compiler is later work.

## Acceptance

- [ ] Samples `testdata/compound_assign*.nra` with `.out`, clean under AddressSanitizer and the leak check: every operator on int and a sized type, string `+=` in a loop, field and index targets, a target with a call in its index that runs once, `x++` and `x--`, a Map index target.
- [ ] Error tests `testdata/errors/compound_assign*` (codes E0962 to E0964): a `const` target, a type mismatch (`s += 1` for a string), `++` on a non-integer, prefix `++x`, a compound assignment used as a value.
- [ ] SPEC.md section 2 no longer says "There are no compound assignments"; the operator list and the "Differences from TypeScript" table match. D76 is amended. The inbox entry about SPEC and `+=` is answered.
- [ ] `sh scripts/test.sh` passes.
