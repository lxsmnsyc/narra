---
id: float-shortest
title: "Float.toString prints 2e+01"
status: todo
priority: must
type: fix
story: batch-15
early: false
outline: false
milestone: p2-std
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
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/float-shortest/**
  - testdata/float_shortest*
shared:
  - selfhost/runtime/*.c
  - selfhost/src/runtime*.nra
  - selfhost/src/stdfloat.nra
  - std/src/ext__float.nra
reserves: []
provides: []
depends_on: []
---
# Float.toString prints 2e+01

## Goal

`Float.toString(20.0)` and `F32.toString` print `2e+01`; SPEC says floats print as the shortest text that reads back. Fix the formatting to the shortest round-trip form, written like JavaScript's `Number.prototype.toString` where SPEC does not say otherwise (`20`, `0.1`, `1e+21`, `1.5e-7`), or as SPEC states if it differs: read the SPEC text first.

## Context

- On main too. Find the C formatter (`grep -rn 'snprintf\|%g' selfhost/runtime`). Shortest round trip: try precision 1..17 (9 for f32) and keep the first that reads back, or an existing Ryu-like routine if one is in the runtime.
- Runtime text is generated: edit the `.c` source and regenerate with `sh scripts/test.sh --suite generated --update`. Template holes and `print` of floats go through the same function: check both.

## Acceptance

- [ ] `testdata/float_shortest.nra` fails on the base and passes: 20.0, 100.0, 0.1, 1/3, 1e21, 1e-7, -0.0, NaN, infinities, `Float.max`, f32 values; each output parsed back with `Float.parse` equals the input.
- [ ] `sh scripts/test.sh` passes (update goldens whose float text changes and list them in `outcome.md`).

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
