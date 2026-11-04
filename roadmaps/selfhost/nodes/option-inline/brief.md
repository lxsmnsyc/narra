---
id: option-inline
title: Store small enum payloads inline
status: todo
priority: later
type: feature
outline: true
milestone: m6-performance
target:
from:
size: L
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-04
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: bench-harness
    kind: hard
    why: needs baseline numbers to show the gain
---

# Store small enum payloads inline

## Goal

Every `Some(x)` heap-allocates a counted payload today, even `Option<int>` (DECISIONS.md: an enum with a fielded variant is `{ tag, payload pointer }`). `None` costs nothing (tag only, NULL pointer), but the emitter still writes a `retain` call on a freshly built `None`. Store payloads of small, non-recursive enums (such as `Option<int>`, `Option<char>`, `Result<int, E>` with small `E`) inline in the enum value, and drop the retain on fieldless constructors. Measure with bench-harness first.

## Notes

Outline. Changes the C layout, so both the Go emitter and `selfhost/` emitter change together, and the D-number decision on enum layout needs an update. Overlaps perceus and layout-tables; refine them together.
