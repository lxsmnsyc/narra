---
id: contract-logic
title: "Logical operators in requires and ensures"
status: todo
priority: should
type: feature
story: batch-11
early: false
outline: true
milestone: 
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-11
spent:
owns:
shared:
reserves: []
provides: []
depends_on:
  - on: array-patterns
    kind: hard
    why: D102 question 4 (`&&` or a range inside an `ensures` pattern) changes the Pattern enum (ast.nra:478-494, PatCompare) and contracts.nra, which array-patterns also changes in every Pattern switch
  - on: contract-logic-design
    kind: hard
    why: builds what D102 decides
---

# Logical operators in requires and ensures

## Goal

Build the contract operators D102 decides (node contract-logic-design): likely `||` and two-parameter comparisons in `requires`, and AND inside an `ensures` pattern, with the prover changes they need. Refine into briefs once D102 is approved.

## Plan review (2026-11-12)

- When refined after the D102 gate, list shared files: ast.nra, contracts.nra, the Pattern switches in the checker, emitter, fmt and astdump.
