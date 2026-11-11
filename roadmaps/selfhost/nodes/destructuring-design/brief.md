---
id: destructuring-design
title: "Spike: destructuring in parameters, array patterns and assignments"
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
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-11
spent:
owns:
  - roadmaps/selfhost/nodes/destructuring-design/**
shared:
  - DECISIONS.md
reserves: [decision D103]
provides: []
depends_on: []
---

# Spike: destructuring in parameters, array patterns and assignments

## Goal

Maintainer request, 2026-11-11: destructuring should reach the places TypeScript allows it. Checked on main (out/narrac, 2026-11-11):

- Works: `const`/`let` struct and tuple destructuring, nested (`const { a: { b }, c } = s`), `match` and `is` patterns, `for (const { c } of items)`.
- Fails: a function parameter (`function area({ width, height }: Rect)`, `function add(#[x, y]: #[int, int])`: "expected a name, found {"), the same in an arrow function, array patterns (`const [a, b] = items`, `match (items) { [] => ..., [first, ...rest] => ... }`: "expected a name, found ["), and destructuring assignment (`#[x, y] = #[y, x]`: "only variables, fields and array elements can be assigned").

Decide, as D103, each of the three with TypeScript's syntax and meaning where Narra's rules allow:

1. **Parameters**, in function declarations, arrow functions and `implement` functions. The bound names are read-only copies (a parameter is already read-only, D81). Open: whether a `mutable` parameter can destructure (likely not, D81, D85); the signature shown in diagnostics, `narrac query` and the language server (the pattern, or a generated name); named arguments (D87) for a destructured parameter; the unused-name rule per bound name; `requires` on a bound name.
2. **Array patterns** in `match`, `is`, and `const`/`let`: `[]`, `[only]`, `[first, ...rest]`. Open: whether `rest` is a copy (D60 slices copy) or a view, and its cost; exhaustiveness over lengths; `const [a, b] = items` outside `match`, where a length mismatch needs a panic (against D92) or a refutable form only (`if (items is [a, b])`); FixedArray patterns, where the length is known and the pattern is irrefutable.
3. **Destructuring assignment**: `#[x, y] = #[y, x]` and `{ x, y } = point` into existing `let` names. Open: evaluation order (all reads before any write, so a swap works), whether targets can be fields and elements, and the shared-write rule (E0978) for each target.

Order the three by value; any can be rejected with a reason. Count uses in `selfhost/src` and `std/src` that each form would shorten.

## Acceptance

- D103 in `DECISIONS.md`, approved by the maintainer (gate).
- Refined briefs for param-destructuring, array-patterns and destructuring-assign, or a reason to drop each.
