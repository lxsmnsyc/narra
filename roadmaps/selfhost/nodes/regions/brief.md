---
id: regions
title: "Regions"
status: todo
priority: must
type: feature
story: batch-13
early: false
outline: false
milestone: p1-syntax
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/regions/**
shared:
reserves: []
provides: []
depends_on:
  - on: regions-plan
    kind: hard
    why: the design gate fixes the syntax, codes and SPEC text this node builds
---
# Regions

## Goal

Build D100 and D113: `Region.run` with bump allocation and copy-out. The plan node splits this node into children.

## Context

- Refined by `regions-plan` after its gate. Read its decision record and outcome first; they override this brief where they differ.
- Proof first: the feature's sample program fails to compile on the batch base.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check.
- [ ] Error tests for each code the decision gives this node.
- [ ] fmt, astdump, the language service and `narrac query` cover any new syntax; the `sources` suite passes.
- [ ] `compiler/README.md` updated.
