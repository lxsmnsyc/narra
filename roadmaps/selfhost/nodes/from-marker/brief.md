---
id: from-marker
title: "The from result marker"
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
    why: same
---

# The from result marker

## Goal

`from <parameter>` after the result type, part of the function type (P10, D94). Body check: returning a value that shares a parameter without the marker is an error with a fix. Function-type assignability: fresh fits `from`, not the reverse. Settle closures returning captures and generic results. Printed in errors and `narra query`.

## Notes

Outline, from D94 and `roadmaps/selfhost/debate/value-semantics/debate.md` (step 2 of 6). Refine with the debate's claims named here.
