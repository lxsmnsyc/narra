---
id: checker-parity
title: Self-hosted diagnostics match every Go diagnostic
status: todo
priority:
type: fix
story: c-seed
early: false
outline: true
milestone:
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-08
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: seed-parity-spike
    kind: hard
    why: the mismatch list from the spike is this node's scope
---

# Self-hosted diagnostics match every Go diagnostic

## Goal

Close every gap the spike finds between the self-hosted checker and the Go checker: diagnostic codes, messages and positions on `compiler/testdata/errors/` and `diagnostics/`, and anything else the spike lists. Outline until the spike reports; refine into one or more briefs from its mismatch list.
