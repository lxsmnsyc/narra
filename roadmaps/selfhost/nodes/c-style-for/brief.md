---
id: c-style-for
title: "C-style for loops"
status: todo
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
added: 2026-11-04
spent: 154436
owns:
  - testdata/c_style_for*
  - testdata/errors/c_style_for*
  - roadmaps/selfhost/nodes/c-style-for/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [code E0959, code E0960, code E0961]
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes go into the Narra compiler only, after the bootstrap fixed point
---

# C-style for loops

## Goal

DECISIONS.md D76 (maintainer, 2026-11-04): `for (init; cond; step) { ... }` and `for (;;)` with TypeScript's meaning. Today both fail with "semicolons are not used", because SPEC.md section 2 and D45 ruled them out.

Rules:

- `init` is a `let` declaration or empty. Its name is scoped to the loop. `const` in `init` is an error (the step could not change it).
- `cond` has type `bool` (no truthiness) or is empty, which loops forever.
- `step` is an assignment statement, a call, or empty. Once compound-assign lands, `i += 1` and `i++` work there too; this node does not depend on it.
- Semicolons are allowed only inside these parentheses.
- `continue` runs the step, then the condition. Labels apply as for other loops.
- The narrowing-end rule for loops that assign (let-narrowing) applies to names the step assigns.
- Not a new kind of value: the loop is a statement.

New AST node kind: add the checker, emitter, `fmt` and `astdump` cases, and a `narrac fmt` layout (`for (let i = 0; i < n; i = i + 1) {`, as oxfmt prints it).

Seed rule: do not use the new syntax in `selfhost/src` in this batch. The seed is regenerated at the end of the batch; using it in the compiler is later work.

## Acceptance

- [ ] Samples `testdata/c_style_for*.nra` with `.out`, clean under AddressSanitizer and the leak check: a counting loop, `for (;;)` with `break`, `continue` that runs the step, a labeled nested loop, an `init` name reused after the loop.
- [ ] Error tests `testdata/errors/c_style_for_*` (codes E0959 to E0961): a non-`bool` condition, `const` in `init`, a semicolon outside the parentheses still reports "semicolons are not used", and use of the `init` name after the loop.
- [ ] `narrac fmt` keeps the loop on one line when it fits; fmt tests cover it.
- [ ] SPEC.md section 2 describes the loop; the "Differences from TypeScript" table and D45 are updated; D76 loses "not built yet" for this half.
- [ ] `sh scripts/test.sh` passes.
