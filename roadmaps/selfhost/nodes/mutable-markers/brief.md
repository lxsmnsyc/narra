---
id: mutable-markers
title: "Where a mutable marker can bind a reference to a part of a value"
status: todo
priority: should
type: spike
story:
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: S
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-07
spent:
owns:
  - roadmaps/selfhost/nodes/mutable-markers/**
shared:
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [decision D85]
provides: []
depends_on:
  - on: mutable-reference
    kind: hard
    why: builds on D81, where bindings that copy part of a value are read-only
---

# Where a mutable marker can bind a reference to a part of a value

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

The maintainer (2026-11-07), after choosing read-only copying bindings for D81: "add a node where we'll explore where we can insert the mutable indicator. For example, `if (x is Some(mutable y))`."

Under D81, `match`/`is` pattern bindings, `for ... of` bindings and destructured bindings are read-only copies. A `mutable` marker on such a binding would make it a reference to that part of the source instead, so writes through it change the source, as TypeScript readers expect. Find every position where the marker could go, decide which ones Narra gets, write D85 and the SPEC text, and stop at the gate. A build node follows.

## Candidates

- Pattern bindings: `if (x is Some(mutable y))`, `match (shape) { Shape.Circle(mutable c) => ... }`, nested patterns.
- Loops: `for (mutable p of players) { p.hp = 0 }`, and `for (const #[index, mutable p] of Arrays.entries(players))`.
- Destructuring: `const { mutable position } = player`, `const #[mutable a, b] = pair`.
- A local reference binding: `mutable frame = emitter.frames[index]` (parked in the inbox from the first draft of mutable-reference), and `x = f(mutable x)` with it.
- Any other place the spike finds (closure parameters, `while` conditions with `is`).

## Questions to answer for each candidate

- What the reference points at, and what the source must be: a `const` or `let` name or a path into one (as a `mutable` argument), never a temporary (`for (mutable p of makePlayers())`).
- Borrow rules: while the reference lives, the source cannot be used through an overlapping path (E0319 and E0929 style), and the reference cannot escape (E0915 style). For loops: can the loop body touch the array (push, reassign) while `p` points into it? A realloc would leave `p` dangling.
- Element types: a reference to an enum payload (`Some(mutable y)`) changes the payload in place. Check that this does not conflict with D81 rule 7 (a struct inside an enum payload cannot change in place); say which wins.
- Copy on write: the source array or payload must be made unique before the reference is taken, as for a `mutable` argument (D50).
- How it lowers in C (a pointer local, as `Local.ByRef`), and the cost in both compilers.
- How many places in `selfhost/src` would use each form (count repeated `mutable path[...]` arguments and index loops that write `items[index].field`).

## Acceptance

- [ ] DECISIONS.md D85: the chosen positions, the rejected ones with reasons, borrow rules, prior art (Rust `ref mut` patterns and `iter_mut`, Swift `inout` and its lack of reference bindings, Hylo `inout` bindings, C# `ref` locals and `foreach (ref var x in span)`), and cost.
- [ ] SPEC sections 2 and 6 describe the chosen forms, marked not built yet in `compiler/README.md`.
- [ ] The outcome has the counts and a short plan for the build node (Go and selfhost sites to change, new diagnostic codes needed).
- [ ] The inbox entry about `mutable name = place` and `x = f(mutable x)` is answered in D85 (chosen, rejected or deferred).

## Sizing

S. Documents and counts.
