---
id: contract-logic
title: "Logical operators in requires and ensures"
status: todo
priority: should
type: feature
story:
early: false
outline: true
milestone: m12-spec-gaps
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
  - on: contract-logic-design
    kind: hard
    why: builds what D102 decides
---

# Logical operators in requires and ensures

## Goal

Build the contract operators D102 decides (node contract-logic-design): likely `||` and two-parameter comparisons in `requires`, and AND inside an `ensures` pattern, with the prover changes they need. Refine into briefs once D102 is approved.
