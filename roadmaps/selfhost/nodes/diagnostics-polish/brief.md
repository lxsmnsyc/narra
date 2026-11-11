---
id: diagnostics-polish
title: "Diagnostics: typeof cycles, E0982, a narrowing cascade, long lines"
status: todo
priority: should
type: fix
story: batch-9
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
  - roadmaps/selfhost/nodes/diagnostics-polish/**
shared:
  - selfhost/src/resolve.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/check.nra
  - selfhost/src/effects.nra
  - testdata/**
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# Diagnostics: typeof cycles, E0982, a narrowing cascade, long lines

## Goal

Four inbox items, each a wrong or misleading message:

1. From the type-of review: for a mutual cycle (`function f1(x: typeof g1)`, `function g1(x: typeof f1)`), E0981 is reported at f1 but says "typeof g1 refers to its own declaration". Name the cycle (`f1 -> g1 -> f1`) or report at the declaration being resolved (`reportTypeofCycle` in `selfhost/src/resolve.nra`).
2. Same review: E0982 also means "type not known yet" for an unannotated same-module constant in a signature. Give that case a code of its own, next free code in the diagnostics table.
3. From the return-trailer review: `testdata/errors/narrowing_shadowed_panic.err` records "expected int, found void" after the narrowing error. A field read on a failed narrowing should not report a second error.
4. From effect-int-keys: `packPosition` in `selfhost/src/effects.nra` packs line and column into 20 bits each (`KEY_STRIDE = 1048576`). A line or column at or past that silently aliases another position. Report an error for such a source (or widen the packing if it costs nothing measurable).

## Acceptance

- Each item has an error case whose `.err` shows the old output on main (record in `log.md`) and the new one on the branch.
- `sh scripts/test.sh` passes.
