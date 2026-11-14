---
id: parameter-defaults
title: "Parameter defaults"
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
  - roadmaps/selfhost/nodes/parameter-defaults/**
shared:
reserves: []
provides: []
depends_on:
  - on: parameter-defaults-design
    kind: hard
    why: the design gate fixes the syntax, codes and SPEC text this node builds
---
# Parameter defaults

## Goal

Build D115, or close as dropped if D115 rejects defaults.

## Context

- Refined by `parameter-defaults-design` after its gate. Read its decision record and outcome first; they override this brief where they differ.
- Proof first: the feature's sample program fails to compile on the batch base.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check.
- [ ] Error tests for each code the decision gives this node.
- [ ] fmt, astdump, the language service and `narrac query` cover any new syntax; the `sources` suite passes.
- [ ] `compiler/README.md` updated.
