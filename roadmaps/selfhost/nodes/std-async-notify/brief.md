---
id: std-async-notify
title: "std/async: a wake-up primitive for tasks, and Net.close"
status: todo
priority: could
type: feature
story:
early: false
outline: true
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent:
owns:
  - roadmaps/selfhost/nodes/std-async-notify/**
shared: []
reserves: []
provides: []
depends_on: []
---

# std/async: a wake-up primitive for tasks, and Net.close

## Goal

Tasks on one loop can wait for each other without polling or tricks: a notify or async mutex/queue in `std/async`, and an explicit `Net.close(connection)` that ends a socket at once and wakes its waiting reader and writer.

Found by `pkg-websocket` (2026-11-26): with no task-level lock or condition, the package built a private wake-up from a day-long sleep that a waker cancels (`packages/websocket/src/wakeup.nra`), and with no close function a connection ends only when every copy is dropped. Refining this node picks the primitive (notify, async mutex, or an async bounded channel), then moves websocket onto it.
