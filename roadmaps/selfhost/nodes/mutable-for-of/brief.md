---
id: mutable-for-of
title: "for (mutable item of items) changes array elements in place (D85)"
status: todo
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
added: 2026-11-09
spent:
owns:
  - testdata/mutable_for*
  - testdata/errors/mutable_for_*
  - roadmaps/selfhost/nodes/mutable-for-of/**
shared:
  - selfhost/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [code E0954, code E0955, code E0956, code E0957, code E0958]
provides: []
depends_on:
  - on: mutable-markers
    kind: hard
    why: D85 fixes the forms and rules this node builds
  - on: mutable-bindings
    kind: hard
    why: builds on D81's split of reassignable names and changeable contents in the checker
---

# for (mutable item of items) changes array elements in place (D85)

## Goal

Build the loop form D85 chose: `for (mutable item of items)`, the same over `Arrays.values(items)`, and `for (const #[index, mutable item] of Arrays.entries(items))`. The binding is a pointer to the element and follows the rules of a `mutable` parameter. The source is a local place of type `Array<T>` with `T` a contents type; the body may not use the source or any path that overlaps it. Not allowed in `function*` or `async` bodies.

The build plan, the selfhost sites and two catches (the loop's own reference to the array, and the parser's pattern path that copies) are in `nodes/mutable-markers/outcome.md`.

## Acceptance

- [ ] Samples `testdata/mutable_for*.nra` with `.out`, clean under AddressSanitizer and the leak check: element writes through the binding, nested fields, the entries form, an array shared before the loop (made unique, the other holder unchanged).
- [ ] Error tests `testdata/errors/mutable_for_*` for: a source that is not a local place, an element type that is not a contents type, a use of the source in the body, the form in a generator or async body, and a write to a `for (const ...)` binding that suggests the marker form.
- [ ] compiler/README.md no longer lists the form as not built; SPEC.md matches.
- [ ] `sh scripts/test.sh` passes.
