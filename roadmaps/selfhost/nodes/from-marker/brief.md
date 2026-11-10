---
id: from-marker
title: "The from result marker"
status: active
priority: should
type: feature
story: batch-5
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
  - testdata/from_*
  - testdata/errors/from_*
  - testdata/fmt/*from_marker*
  - roadmaps/selfhost/nodes/from-marker/**
shared:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - testdata/**
  - docs/guide/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: [diag E0976, diag E0977]
provides: []
depends_on:
  - on: copy-builtin
    kind: hard
    why: the debate orders them one after the other (P20); both edit the checker's sharing facts
---

# The from result marker

## Goal

Build the `from <parameter>` marker from D94 (debate `roadmaps/selfhost/debate/value-semantics/debate.md`, settled structure item 3, claims P8, P10, O13).

- Syntax: `function mergeSort(items: Array<SortItem>): Array<SortItem> from items`. The marker follows the result type and names one parameter. It is part of the function type, like effects (D86): `(grid: Grid) => Array<int> from grid`.
- Body check (E0976): returning a value that shares a parameter without `from` naming it is an error, checked the way D46 checks `escaping`. The fix adds `from <parameter>`. Nothing is inferred across functions.
- Assignability (E0977): a function with fresh results fits a slot with `from`; a function with `from` does not fit a slot without it. The fix adds the marker to the slot's type.
- Closures (settled 2026-11-10, the maintainer may still revise): `from` can name only a parameter, never a capture. An arrow function whose result shares a captured value is an E0976 error; the fix wraps the returned value in `copy(...)`. An arrow function whose result shares one of its own parameters writes `from <parameter>` like any function.
- Generic results (settled 2026-11-10): `from items` covers the whole parameter and everything reached through it, at any depth, whatever the type arguments are. A result of type `T` taken out of `items: Array<T>` shares `items`. When `T` is a scalar type the marker is still allowed and changes nothing.
- Record both answers in DECISIONS.md as an amendment to D94 (no new number).
- The marker is printed in errors, hover text and `narra query` signatures.

The compiler's own functions get their marks in shared-write-sweep, not here.

## Acceptance

- [ ] `from` parses in declarations and function types, and formats, dumps and round-trips through `narrac fmt`.
- [ ] E0976 and E0977 each have an error sample whose fix is applied by `narrac apply-fix`.
- [ ] A sample passes a fresh function to a `from` slot and runs.
- [ ] SPEC.md section 6 describes the marker, and DECISIONS.md D94 records the closure and generic answers.
- [ ] Error samples: an arrow function returning a captured array (E0976, fix `copy(...)`), and a generic function returning an element of its `from` parameter that passes the check.
- [ ] `sh scripts/test.sh` passes.
