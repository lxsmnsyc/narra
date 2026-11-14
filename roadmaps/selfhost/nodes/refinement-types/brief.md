---
id: refinement-types
title: "Refinement types on distinct types"
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
size: M
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-13
spent:
owns:
  - roadmaps/selfhost/nodes/refinement-types/**
shared:
reserves: []
provides: []
depends_on:
  - on: refinement-types-design
    kind: hard
    why: the design gate fixes the syntax, codes and SPEC text this node builds
---
# Refinement types on distinct types

## Goal

Build D116: refinement types on distinct types.

## Context

- Refined by `refinement-types-design` after its gate. Read its decision record and outcome first; they override this brief where they differ.
- Proof first: the feature's sample program fails to compile on the batch base.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check.
- [ ] Error tests for each code the decision gives this node.
- [ ] fmt, astdump, the language service and `narrac query` cover any new syntax; the `sources` suite passes.
- [ ] `compiler/README.md` updated.
