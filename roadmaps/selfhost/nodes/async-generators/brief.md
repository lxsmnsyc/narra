---
id: async-generators
title: "Async generators and for await"
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
  - roadmaps/selfhost/nodes/async-generators/**
shared:
reserves: []
provides: []
depends_on:
  - on: async-generators-design
    kind: hard
    why: the design gate fixes the syntax, codes and SPEC text this node builds
---
# Async generators and for await

## Goal

Build D110: `async function*`, `AsyncGenerator<T>` and `for await`.

## Context

- Refined by `async-generators-design` after its gate. Read its decision record and outcome first; they override this brief where they differ.
- Proof first: the feature's sample program fails to compile on the batch base.

## Acceptance

- [ ] Samples with checked output, clean under AddressSanitizer, the leak check and the copy check.
- [ ] Error tests for each code the decision gives this node.
- [ ] fmt, astdump, the language service and `narrac query` cover any new syntax; the `sources` suite passes.
- [ ] `compiler/README.md` updated.
