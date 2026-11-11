---
id: unsafe-ptr
title: "Unsafe code: unsafe blocks, Ptr<T> and @trusted"
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
spent: 581821
owns:
  - roadmaps/selfhost/nodes/unsafe-ptr/**
  - selfhost/src/unsafecode.nra
  - selfhost/src/emitpointers.nra
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
  - selfhost/src/effects.nra
  - selfhost/src/types.nra
  - selfhost/src/builtins.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/runtimetext.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitstmts.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/emit.nra
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
  - selfhost/src/exprs.nra
  - selfhost/src/buildconditions.nra
reserves: [diag E0996, diag E0997, diag E0998]
provides: []
depends_on:
  - on: attributes
    kind: hard
    why: @trusted needs the general attribute list
  - on: const-generics
    kind: hard
    why: both add Type variants in the same key and C-name functions (types.nra buildKey, cnames.nra typeId); const-generics changes that plumbing first
---

# Unsafe code: unsafe blocks, Ptr<T> and @trusted

## Goal

D95 C interop, step 1 part 1: `unsafe { ... }` is the only place pointer operations and unsafe calls are allowed. Pointers are values of type `Ptr<T>` with functions, not operators: `Ptr.read`, `Ptr.write`, `Ptr.offset`, `Ptr.alloc`, `Ptr.free`. A function that contains an `unsafe` block has the `unsafe` effect unless it is marked `@trusted`, which stops the effect at a hand-checked wrapper. Code inside `unsafe` may crash; it is the stated exception to D92.

## Context

- Design: D95 "C interop". It replaces SPEC section 9's first bullet; rewrite that bullet (keep the parts step 2 owns, `@layout` and C-callable functions, as "not built yet").
- `unsafe` is already a known effect name (`effects.nra`, `KNOWN_EFFECTS`).
- Attributes: node attributes makes them a general list; add `@trusted` as a name with its check.
- `Ptr<T>` is a plain value (no reference count). What `T` may be (numbers, `u8`, structs without heap fields?) is a design call to state in SPEC; heap types behind a raw pointer would bypass reference counting.

## Acceptance

- [ ] Samples: `Ptr.alloc`/`write`/`read`/`offset`/`free` in an `unsafe` block, the `unsafe` effect required on the caller, a `@trusted` wrapper callable without it, pointer use outside `unsafe` (error), `@trusted` on a function with no `unsafe` block (error or warning, say which), and the debug leak check passing for a program that frees what it allocates.
- [ ] Fast check passes: `sh scripts/test.sh`
