---
id: param-destructuring
title: "Destructuring in parameters"
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

# Destructuring in parameters

## Goal

Struct and tuple patterns as parameters of functions, arrow functions and `implement` functions (`function area({ width, height }: Rect)`, `(#[x, y]: #[int, int]) => x + y`), as D103 decides. Parser, checker, emitter, `fmt`, `astdump`, diagnostics and `narrac query` signatures. Inbox item of 2026-11-07, approved for planning. Refine into a full brief once D103 is approved.
