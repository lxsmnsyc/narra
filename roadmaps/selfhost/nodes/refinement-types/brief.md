---
id: refinement-types
title: "Refinement types on distinct types"
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

# Refinement types on distinct types

## Goal

`distinct type Port = int requires >= 1 && <= 65535`, `distinct type Method = string requires "GET" | "POST"`. `Port.from(raw): Option<Port>` checks at runtime; the prover knows the predicate inside. `Json.decode` of such a type returns Err when the check fails. Uses the D95 and D102 contract patterns. From the D95 discussion, 2026-11-09. Refine through a design spike.

## Notes

Outline, added 2026-11-13 when the maintainer set the phase order (syntax, std, feature review, packages, distribution, docs).
