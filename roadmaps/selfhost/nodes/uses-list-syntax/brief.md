---
id: uses-list-syntax
title: "Effect lists inside type lists"
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
  - roadmaps/selfhost/nodes/uses-list-syntax/**
shared:
reserves: []
provides: []
depends_on:
  - on: uses-list-design
    kind: hard
    why: the design gate fixes the syntax, codes and SPEC text this node builds
---
# Effect lists inside type lists

## Goal

Build D117: the new effect list syntax inside type lists.

## Context

- Refined by `uses-list-design` after its gate. Read its decision record and outcome first; they override this brief where they differ.
- Proof first: the feature's sample program fails to compile on the batch base.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check.
- [ ] Error tests for each code the decision gives this node.
- [ ] fmt, astdump, the language service and `narrac query` cover any new syntax; the `sources` suite passes.
- [ ] `compiler/README.md` updated.
