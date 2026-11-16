---
id: restrict-renames
title: "A second name for a value that changes in place is an error (E0971)"
status: done
priority: should
type: feature
story: batch-18
early: false
outline: false
milestone: p3-review
target:
ship: auto
land: auto
from:
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-16
spent: 549151
owns:
  - selfhost/src/check/sharedwrites.nra
  - selfhost/src/driver/diagnostics.nra
  - selfhost/src/driver/fixes.nra
  - testdata/**
  - SPEC.md
  - docs/guide/**
  - website/src/home/**
  - roadmaps/selfhost/nodes/restrict-renames/**
shared:
  - selfhost/src/**
  - std/src/**
  - seed/narrac.c
reserves: [diag E0971]
provides: []
depends_on:
  - on: value-copy-semantics
    kind: hard
    why: the debate settled the rule
---
# A second name for a value that changes in place is an error (E0971)

## Goal

Build D124 (DECISIONS.md; debate `roadmaps/selfhost/debate/value-copy/debate.md`). `const b = a`, `let b = a` and `b = a`, where `a` is a bare name (local, parameter, pattern or loop binding, top-level constant) of a type D94 tracks (struct including all-scalar, `Array`, `Map`, `FixedArray`, `Box`, `StringBuilder`, generator, coroutine), are E0971 at the binding:

    synonym would be a second name for the array in prices; use prices, or copy(prices) for an array of its own

with an `apply-fix` that wraps the right side in `copy(...)`. Renames of scalars, strings, bytes, enums, tuples and function values stay allowed. Field and element reads, arguments, results, literals and captures keep E0978 and E0976 unchanged. E0971 is free (its old deprecation use was removed in batch 3).

## Steps

- The check, in the shared-write walk, so it runs wherever E0978 runs (every command that checks a package and the language server). A rename that E0971 reports is not reported again as E0978.
- Rewrite the compiler and std sites (debate fact J15: about 13 typed sites in `selfhost/src`, all a default reassigned later such as `let ordered = arguments` in `check/calls.nra`, plus untyped std sites the check finds). Prefer an `if` or `match` expression or a helper over `copy(...)` where no second value is meant. Regenerate the seed once the compiler's sites change (`sh scripts/seed-update.sh`).
- Error samples for each binding form and each tracked type, and allowed samples for the untracked types. Update samples under `testdata` that the rule now rejects.
- SPEC section 6: a "Renames" paragraph before "Shared writes"; the E0978 text no longer lists a bare-name binding as a sharing start. Guide chapter 2, chapter 9 (`let backup = hero` and the E0978 explanation), chapter 13 (`const before = scores`, twice). Home section 02 already uses `copy(hero)`; check it still reads right.

## Acceptance

- [ ] The maintainer's snippet is E0971 at `const synonym = prices` with the message above, and `apply-fix` makes it compile.
- [ ] `sh scripts/test.sh` passes; the seed builds the compiler; the website build passes.
- [ ] D124's "Not built yet" line is replaced with what was built.
