---
id: std-no-panic
title: "Std functions return Result or Option instead of panicking"
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
added: 2026-11-09
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: batch-2
    kind: hard
    why: builds on the compound assignment, C-style for and std modules batch-2 added
---

# Std functions return Result or Option instead of panicking

## Goal

D92: `Str.slice`, `Str.split`, `Str.repeat`, `Bytes.at`, `Bytes.slice`, `Float.truncate`, `Arrays.filled`, `Ranges.stepped`, `Math.clamp`, `Math.powInt` and `Random.nextInt` stop panicking on bad input. Pick `Option` where there is one failure and `Result` where the caller needs the reason; name them in a table in SPEC.md. Remove compound assignment on a Map index target (no special error for the removed form; GOAL.md constraint) and add `Maps.update(mutable map, key, initial, change)`. Update every caller in `selfhost/src` and `std/src`, the guide and the samples.

## Notes

Outline, from D92 (maintainer, 2026-11-09). `selfhost/src` must follow the new rule in the same node: fix every site it rejects. Update SPEC.md (sections 2, 4 and 6 as needed) and drop "Not built yet" from D92 for this part.
