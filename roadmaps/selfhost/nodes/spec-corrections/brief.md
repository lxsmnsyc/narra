---
id: spec-corrections
title: "SPEC.md matches the compiler"
status: done
priority: could
type: docs
story: batch-21
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: pnpm --dir website build
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/spec-corrections/**
shared:
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: guide-targets
    kind: hard
    why: collects what the last chapter writers find; runs near the end of the batch
---
# SPEC.md matches the compiler

## Goal

The guide writers of batch-21 compare SPEC.md with the compiler. Fix what they found, so the spec page a reader sees on the site is correct. Found so far:

- Section 8: the sample threads (`firstHalf`, `secondHalf`) do not compile, because `%` by a name not checked against zero is rejected; add the check (guide-threads).
- `Channel.values` is listed as `uses io` only; looping over it also needs `alloc` (guide-threads).
- Section 9: `return` inside an `unsafe` block works but the spec does not say so (guide-c-interop).

- Section 2 says "An iterator is a `Generator<T>`", but `Maps.keys`, `Maps.entries` and `Str.chars` return arrays, `Arrays.values` and `Arrays.entries` are allowed only at the head of a loop (E0920), and a range has its own type `Range`; chapter 12's table "What a loop can read" teaches what the compiler does (guide-generators).

- A struct literal does not work out a constant type parameter from its fields (`Schedule { hours: [...] }` fails, `Schedule<14> { ... }` works); the spec says only that calls work it out. Say so (guide-proofs).

- Section 3 lists `narrac move old new` (move a file and rewrite its imports), but the compiler has no such command. Move the line to section 16 (Deferred) as not built yet (guide-tools-gaps).

Before editing, read every batch-21 node `outcome.md` for more "spec against compiler" findings. Compile every sample of the sections you touch with a compiler built from this branch; fix the spec text, not the compiler.

## Acceptance

- [ ] Each finding fixed in SPEC.md, checked with the compiler.
- [ ] `pnpm --dir website build` passes.
