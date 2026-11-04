---
id: c-style-for
title: "C-style for loops"
status: todo
priority: should
type: feature
outline: true
milestone: m12-spec-gaps
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes land in the Go compiler and selfhost/ together, after the bootstrap fixed point
---

# C-style for loops

## Goal

DECISIONS.md D76 (maintainer, 2026-11-04). `for (init; cond; step) { ... }` and `for (;;)` with TypeScript's meaning. Today both fail with "semicolons are not used", because SPEC.md:121 and D45 ruled them out.

## Notes

- Independent of compound-assign: the step works as `i = i + 1` before `i += 1` or `i++` lands.
- Outline. Rules to settle when refining: `init` is a `let` declaration (or empty), scoped to the loop; `cond` is `bool` (no truthiness) or empty for an endless loop; `step` is an assignment statement (or empty); semicolons are allowed only inside these parentheses; `continue` runs the step; labels (labels node) and the narrowing-end rule for loops that assign (let-narrowing) apply; the tail-loops pre-pass and block-expression boundaries still hold. Both compilers, samples under ASan, error tests, SPEC.md:121 and the "Differences from TypeScript" table.
- Then use it in selfhost/src where `while` loops carry a manual step, if that reads better.
