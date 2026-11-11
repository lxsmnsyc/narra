---
id: array-patterns
title: "Array patterns in match, is and const"
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
  - on: destructuring-design
    kind: hard
    why: builds what D103 decides
---

# Array patterns in match, is and const

## Goal

Array patterns `[]`, `[only]`, `[first, ...rest]` in `match` and `is`, with exhaustiveness over lengths, and in `const`/`let` as D103 decides (refutable or not), including FixedArray patterns. Inbox item of 2026-11-07. Refine into a full brief once D103 is approved.
