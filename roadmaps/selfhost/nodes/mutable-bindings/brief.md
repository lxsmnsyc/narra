---
id: mutable-bindings
title: "Build D81: mutable changes contents, const fixes only the name"
status: todo
priority: should
type: feature
story: batch-1
early: false
outline: false
milestone: m12-spec-gaps
target:
ship: pr
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - testdata/mutable_contents*
  - testdata/errors/mutable_contents_*
  - roadmaps/selfhost/nodes/mutable-bindings/**
shared:
  - selfhost/src/**
  - testdata/**
  - docs/guide/**
  - compiler/README.md
  - SPEC.md
reserves: [code E0948, code E0949, code E0950, code E0951, code E0952, code E0953]
provides: []
depends_on:
  - on: mutable-reference
    kind: hard
    why: the gate fixes the rules (D81) this node builds
  - on: closure-borrow
    kind: hard
    why: stack closures capture a reference binding by reference, through the capture mode and conflict scan closure-borrow adds
---

# Build D81: mutable changes contents, const fixes only the name

> 2026-11-09: the Go compiler and Go test harness were retired (story c-seed, PR #101). Build only in `selfhost/src` (and `std/src`, `selfhost/runtime/`); ignore the Go half of any plan below. Go test names below map to `selfhost/tests` suites; run `sh scripts/test.sh`. Refine this brief before starting.

## Goal

Build D81 (DECISIONS.md) in both compilers, and rewrite every program, test and guide chapter it changes. All rules 1 to 10 and the read-only copying bindings are decided by the maintainer (2026-11-07). `nodes/mutable-reference/outcome.md` has file:line lists for every site; read it first, with D81 and SPEC sections 2 and 6.

## Rules to build

1. A `mutable` parameter's type must be a contents type (struct, opaque struct, `Array`, `FixedArray`, `Map`, `Box`, `StringBuilder`, `Generator`, `Coroutine`); anything else is an error at the parameter.
2. No assignment to a whole `mutable` parameter.
3. A local `const` can have fields and elements assigned and be passed as `mutable`; `let` only adds reassignment of the name. Top-level `const` stays fixed at every depth (E0290).
4. Plain parameters and bindings that copy part of a value (`match`/`is` patterns, `for ... of`, destructuring) are read-only: no field or element writes, not passable as `mutable`. The error for a write points at the source (`players[index].hp = 0`) or `const copy = p`.
5. Generics: bare `mutable value: T` is an error; `mutable Self` in a trait requires each `implement` to be for a contents type; `Dynamic<Trait>` cannot go to a `mutable Self` function.
6. Closures: a non-escaping closure that changes a captured local `const`, directly or through a nested closure, captures it by reference in every closure of the chain, and E0614 borrows it for the call. An escaping closure cannot change a captured value.
7. E0319, E0929, E0609 apply to a `const` passed as `mutable`.

## Context

- The outcome's "Gotchas" lists the checker and emitter sites: `Local.Mutable` must split into "can be reassigned" and "contents can change" (`check_stmt.go:331-356`, `check_call.go:604-616`, `check_call.go:665` `checkBorrowed`, `check_closure.go:142`, `emit_closure.go:155-185`), and the selfhost counterparts (`exprs.nra:1714`, `:1769`, `:1779`, `stmts.nra:351`).
- E0227, E0315 and E0290 texts say "declare it with `let`" or "copy it into a `let`"; reword for the new model.
- The unused-names rule "a `mutable` parameter that is only assigned counts as used" (README) becomes "only written through".

## Acceptance

- [ ] Rules 1 to 7 in Go and selfhost with identical diagnostics; new codes from the reserved E0948 to E0953 only.
- [ ] `selfhost/src`: the 3 `int` `mutable` parameters (`patterns.nra:358`, `declsdump.nra:118`, `:267`) and the whole assignment in `modules.nra:237` rewritten as the outcome proposes. Bootstrap fixed point holds.
- [ ] `compiler/testdata`: the 34 fixtures with rejected types and the 28 whole assignments rewritten so each still tests its own rule; `box_assign_const.nra` and the `err_arrays_maps` dump line updated. New error tests `errors/mutable_contents_*.nra`, one per rule, and a run sample `mutable_contents.nra` (`const` struct passed as `mutable`, `const` array pushed, nested closure changing a `const`, ASan and leak check clean).
- [ ] `docs/guide` chapters 2, 3, 9 and 11 rewritten at the lines the outcome lists, teaching the new model with tested samples (no SPEC or D-number citations).
- [ ] `compiler/README.md` drops the D81 "not built yet" line and rewords the unused-names line.
- [ ] Fast check passes, including `TestSelfhostBootstrap`, and the Linux check (`scripts/linux-check.sh`) if Docker runs.

## Sizing

L. Checker split of `Local.Mutable` in both compilers plus a large mechanical rewrite of tests and guide samples. If the worker finds it too large for one reviewable PR, split after the checker change: the rules land with the test rewrites, the guide in a second commit of the same PR.
