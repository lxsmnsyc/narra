---
id: uses-list-syntax
title: "Effect lists inside type lists"
status: todo
priority: must
type: feature
story:
early: false
outline: true
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on: []
---

# Effect lists inside type lists

## Goal

Outside a parameter list, a `uses` list on a type takes every following `, name` as another effect (`parseUsesList`), so `Map<Generator<int> uses io, int>` parses as one generator with effects `io, int`, and `#[Generator<int> uses io, Array<int>]` is a parse error. Function types with `uses` have the same problem. Needs a syntax decision (for example parentheses around the effect list), then the parser, fmt and SPEC change. From the PR #82 review.

## Notes

Outline, added 2026-11-13 when the maintainer set the phase order (syntax, std, feature review, packages, distribution, docs).
