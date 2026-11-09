---
id: misleading-hints
title: "Two misleading compiler hints"
status: active
priority: should
type: fix
story: batch-2
early: false
outline: false
milestone: m9-adoption
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
added: 2026-11-09
spent:
owns:
  - docs/guide/**
  - roadmaps/selfhost/nodes/misleading-hints/**
shared:
  - selfhost/src/**
  - std/src/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: language changes go into the Narra compiler only, after the bootstrap fixed point
---

# Two misleading compiler hints

## Goal

Found while writing guide chapters 6 to 10 (2026-11-05):

1. Assigning to a field of a parameter says "declare it with `let`". For a parameter the fix is a `mutable` parameter or returning a new value. After D81 (batch-1) check what the hint says now and make it name the parameter case.
2. A missing `implement` block for a user trait suggests `derives Label`, but only built-in traits can be derived. Suggest an `implement` block for user traits.

Then update the guide's error samples that show these messages (chapters 7 and 9, under `docs/guide/`).

## Acceptance

- [ ] Proof on the base: error tests that show both hints as they are today.
- [ ] Updated expected output in `testdata/errors/` for both cases; built-in traits still suggest `derives`.
- [ ] Guide chapters 7 and 9 show the new messages, and the guide test passes.
- [ ] `sh scripts/test.sh` passes.
