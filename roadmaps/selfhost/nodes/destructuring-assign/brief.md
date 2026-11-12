---
id: destructuring-assign
title: "Destructuring assignment"
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
shared:
reserves: []
provides: []
depends_on:
  - on: destructuring-design
    kind: hard
    why: builds what D103 decides
---

# Destructuring assignment

## Goal

Assignment to existing names through a tuple or struct pattern (`#[x, y] = #[y, x]`, `{ x, y } = point`), all reads before any write, with the shared-write rule per target, as D103 decides. Refine into a full brief once D103 is approved.
