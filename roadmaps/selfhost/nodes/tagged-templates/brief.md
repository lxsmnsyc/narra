---
id: tagged-templates
title: "Tagged template literals"
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
size: M
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

# Tagged template literals

## Goal

`tag`text ${value}`` with TypeScript's meaning: the tag is a function named at the use site, so no call is hidden. Approved for planning 2026-11-07. Open: the tag's signature (an `Array<string>` of parts and a typed list of values, or one call per value type), how values of different types pass without varargs (D12), the tag's effects, and whether std ships tags (JSON, shell quoting). Refine through a design spike with a decision record.

## Notes

Outline, added 2026-11-13 when the maintainer set the phase order (syntax, std, feature review, packages, distribution, docs).
