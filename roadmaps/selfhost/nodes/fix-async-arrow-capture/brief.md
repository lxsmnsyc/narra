---
id: fix-async-arrow-capture
title: "fix: async codegen bugs found by the websocket package"
status: done
priority: should
type: fix
story: batch-26
early: false
outline: false
milestone: p4-packages
target:
ship: auto
land: auto
from:
size: S
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-26
spent: 140000
owns:
  - roadmaps/selfhost/nodes/fix-async-arrow-capture/**
  - testdata/async-arrow-struct*
  - testdata/async-generator-alias*
shared:
  - selfhost/src/emit/*.nra
  - selfhost/src/check/*.nra
reserves: []
provides: []
depends_on: []
---

# fix: async codegen bugs found by the websocket package

## Goal

Two async forms that the websocket package had to work around compile to valid C and run correctly.

## Context

- Found by `pkg-websocket` (2026-11-26, see its `outcome.md` and `log.md`). The package works around both; after the fix the workarounds may stay, and a follow-up can simplify them.
- Bug 1: an async arrow whose parameter or capture is a struct emits `f->v_x = *v_x` in the C, which does not compile.
- Bug 2: a top-level `const` alias of an async generator function (`export const messages = Session.messages`) gives a C name clash.
- Prove each first: a sample under `testdata/` that fails on the current branch. Then fix and keep the samples.

## Acceptance

- [ ] Samples for an async arrow taking a struct parameter, one capturing a struct, and a const alias of an async generator, each failing on the base and passing on the branch, clean under AddressSanitizer and the leak check.
- [ ] `sh scripts/test.sh` passes and `sh selfhost/bootstrap.sh` reaches the fixed point.
