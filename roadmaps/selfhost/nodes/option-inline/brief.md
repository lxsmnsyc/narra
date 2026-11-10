---
id: option-inline
title: "Store payloads of non-recursive enums inline"
status: active
priority: should
type: feature
story: batch-4
early: false
outline: false
milestone: m6-performance
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: [benchmarks]
fulfilled_by:
gate: false
added: 2026-11-04
spent:
owns:
  - testdata/enum_inline_*
  - roadmaps/selfhost/nodes/option-inline/**
shared:
  - selfhost/src/emitdecls.nra
  - selfhost/src/emitexprs.nra
  - selfhost/src/emitmatch.nra
  - selfhost/src/emitoptions.nra
  - selfhost/src/emitstate.nra
  - selfhost/src/emit.nra
  - DECISIONS.md
  - SPEC.md
  - compiler/README.md
reserves: [decision D99]
provides: []
depends_on:
  - on: perceus
    kind: hard
    why: both rewrite the retain and binding code in emitstate.nra and emitmatch.nra; the layout change lands on top of the new ownership rules
---

# Store payloads of non-recursive enums inline

## Goal

An enum whose variants do not reach the enum itself is emitted as a tagged union held by value, `struct E { int32_t tag; union { ... } u; }`, with no heap payload. Recursive enums (an AST) keep the counted payload pointer of D55. `Some(1)`, `Ok(x)` and `Err(e)` stop allocating. Fieldless constructors such as `None` stop emitting a retain call.

## Context

- Today (D55, `DECISIONS.md:465`): an enum with a fielded variant is `struct E { int32_t tag; Q_E* u; }` (`typeDefinition`, `emitdecls.nra:146-166`), with the payload `struct Q_E { int64_t rc; int32_t tag; union {...}; }` (`emitHeapDefinitions`, 168-192), built by `enumAllocation` (410-445), freed by `drop_<id>` (256-280). Retain is `if (v.u) v.u->rc++`. Payload reads are `(x).u->v_V.f_name` (`emitexprs.nra:303`, `emitoptions.nra:132-145`).
- `Option` is a plain generic prelude enum (`declare.nra:84-87`), monomorphized per type, so its layout follows from the rule with no special case.
- `None` is `((E){.tag = N})`, borrowed, and `consume` adds `retain_E(t)` (see `seed/narrac.c:45976-45978`), a runtime no-op.
- D55 rejected inline payloads because layout would depend on the type's shape and the compiler would need two code paths. This node reverses that for non-recursive enums: write D99, which supersedes that part of D55, with the measured gain as its reason. Amend D55 and SPEC section 6 to match.
- An inline enum value is copied like a struct: retain and release recurse into the fields of the active variant (the same shape as struct RC functions in `typeRcFunctions`, `emitdecls.nra:248-312`). An inline enum with no counted fields needs no RC functions at all (`needsRc`, `emitstate.nra:355-381`).
- Recursion test: a variant field reaches the enum through other enums or structs held by value. A path through `Array`, `Map`, `Box` or a function type is already behind a pointer and does not count.

## Steps

- [ ] Write D99 and measure the size of the largest inline enums in `selfhost/src` (a fat `Result` copied by value could cost more than the allocation it saves). If one is large, D99 sets a size limit above which the payload stays on the heap.
- [ ] Recursion test and the inline layout in `emitdecls.nra`, constructors, reads, match, RC functions.
- [ ] No retain on fieldless constructors.
- [ ] Samples `testdata/enum_inline_*`: Option of int, string and struct; Result with counted fields on both sides; a recursive enum next to an inline one; an enum that reaches itself through another enum (stays on the heap); nested matches and copies of inline values with counted fields.
- [ ] Regenerate the seed; the bootstrap fixed point must hold.

## Acceptance

- [ ] `bench/run.sh` before and after for all programs and the compiler row, reported in the outcome.
- [ ] `sh scripts/test.sh` passes under AddressSanitizer and the leak check, and `sh scripts/linux-check.sh` passes.
- [ ] D99 written; D55 and SPEC section 6 updated.

## Sizing

L: blast radius high (every enum in every emitted program), size medium, uncertainty low to medium (a tagged union is the standard C layout).

## Notes for the worker

- Generic code is monomorphized today, so the layout is known at every use. layout-tables (debug builds compiling generics once) is not built; do not design for it beyond noting in D99 that a layout table would carry the enum's size.
- Do not start reuse of payload cells; perceus-reuse does that next, on the heap enums that remain.
