---
id: async-timing-flake
title: "async_api_all_observed depends on timer order under load"
status: active
priority: must
type: fix
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - testdata/async_api_all_observed*
  - roadmaps/selfhost/nodes/async-timing-flake/**
shared: []
reserves: []
provides: []
depends_on: []
---

# async_api_all_observed depends on timer order under load

## Goal

`testdata/async_api_all_observed.nra` (batch-12) fails about once per full check when the load average is above 30, on macOS and Linux: two timer lines print in the other order. It passes alone. Make its output depend only on what D109 fixes (the order of awaited results, cancel and observer rules), not on which of two close timers fires first: wider gaps between timers, or print after `Async.all` resolves, or compare sets. Keep what the sample proves (a handle given to an unfinished `all` is not cancelled through another waiter).

## Acceptance

- [ ] Proof: run the sample 200 times under load (for example 8 parallel loops) on feat/batch-13 and count failures; record the count.
- [ ] Fix: the same 200 runs under the same load give no failure; the sample still fails if the observed-handle rule in `std_async_loop.c` is removed (check by hand and say so).
