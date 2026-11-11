---
id: fixed-arrays
title: "Fixed arrays: FixedArray<T, N> with proven indexes"
status: active
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
spent:
owns:
  - roadmaps/selfhost/nodes/fixed-arrays/**
  - selfhost/src/fixedarrays.nra
  - selfhost/src/emitfixed.nra
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
  - selfhost/src/types.nra
  - selfhost/src/exprs.nra
  - selfhost/src/stmts.nra
  - selfhost/src/operands.nra
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/emittraits.nra
  - selfhost/src/emit.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/stdcopy.nra
  - selfhost/src/sharedwrites.nra
  - selfhost/src/builtins.nra
  - selfhost/src/stdtext.nra
  - std/src/**
  - selfhost/src/cnames.nra
  - selfhost/src/generics.nra
  - selfhost/src/calls.nra
  - selfhost/src/traits.nra
  - selfhost/src/resolve.nra
  - selfhost/src/stdtypes.nra
  - selfhost/src/declsdump.nra
  - selfhost/src/typesdump.nra
  - selfhost/src/sharedresults.nra
reserves: [diag E0994, diag E0995]
provides: []
depends_on:
  - on: const-generics
    kind: hard
    why: needs constant type parameters for the length
---

# Fixed arrays: FixedArray<T, N> with proven indexes

## Goal

D95 fixed arrays, part 2: `FixedArray<T, N>` stored inline like a struct, with no heap allocation. An array literal builds one when the expected type is a `FixedArray`. Indexing never panics: a constant index is checked at compile time, and a variable index must be proven below `N` and not negative by the prover (`if (i < 4)`, `for (let i = 0; i < N; i++)` with `i` unassigned in the body, `Ranges.upTo(N)`). `FixedArrays.get(values, i): Option<T>` covers the rest, plus a length and `for ... of`.

## Context

- Design: D95 "Fixed arrays"; SPEC section 3 and the `copy` and `mutable` rules that already name `FixedArray` (section 6, D81, D94): it is a contents type, writes change it in place, and the shared-write rule (E0978) and `copy(...)` apply.
- Equality, hashing and `Show` follow the element type, as for `Array`.
- Tree shaking: types and helpers are written only when reached (`noteHelper`/`useHelper`, see the tree-shake-types outcome).
- The prover facts per local are in `operands.nra` (see the contracts and proven-operands outcomes).

## Acceptance

- [ ] Samples: literal with an expected type, constant index in and out of range (error), variable index proven by `if`, by a C-style for loop and by `Ranges.upTo(N)`, unproven index (error), `FixedArrays.get`, `for ... of`, a `mutable` parameter, `copy(...)` and a shared write (E0978), equality and `Show`, a generic `sum<const N: int>`; all clean under the leak check and sanitizers.
- [ ] Fast check passes: `sh scripts/test.sh`
