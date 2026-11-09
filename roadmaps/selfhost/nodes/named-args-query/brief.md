---
id: named-args-query
title: narra query reads named arguments
status: active
priority:
type: feature
story: batch-1
early: false
outline: false
milestone:
target:
ship: auto
from:
size: S
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - testdata/query/named_args*
  - roadmaps/selfhost/nodes/named-args-query/**
shared:
  - selfhost/src/query.nra
  - selfhost/src/querysites.nra
  - selfhost/src/querysignature.nra
reserves: []
provides: []
depends_on:
  - on: named-args-check
    kind: hard
    why: query reads the checked match between labels and parameters
---

# narra query reads named arguments

## Goal

`narra query refs` on a parameter lists the call sites that name it, at the label's position. `narra query callers` and the signature output show labelled calls correctly. Renaming a parameter can then find every call that would break.

## Acceptance

- [ ] Query goldens for refs of a parameter with named call sites, and callers of a function called with named arguments.
- [ ] Fast check passes.
