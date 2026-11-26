---
id: fix-async-arrow-capture
title: "fix: async arrows with struct parameters or captures emit invalid C"
status: todo
priority: should
type: fix
story:
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
spent:
owns:
  - roadmaps/selfhost/nodes/fix-async-arrow-capture/**
  - testdata/async-arrow-struct*
shared:
  - selfhost/src/emit/*.nra
  - selfhost/src/check/*.nra
reserves: []
provides: []
depends_on: []
---

# fix: async arrows with struct parameters or captures emit invalid C

## Goal

An async arrow whose parameter or capture is a struct compiles to valid C and runs correctly.

## Context

- Found by `pkg-websocket` (2026-11-26, see its `outcome.md` and `log.md`): such an arrow emits `f->v_x = *v_x` in the C, which does not compile. The package works around it with plain arrows that call named async functions.
- Prove it first: a sample under `testdata/` that fails on the current branch. Then fix the emitter and keep the sample.
- After the fix, the websocket workaround may stay; a follow-up can simplify it.

## Acceptance

- [ ] A sample with an async arrow taking a struct parameter, and one capturing a struct, fails to compile on the base and passes on the branch, clean under AddressSanitizer and the leak check.
- [ ] `sh scripts/test.sh` passes and `sh selfhost/bootstrap.sh` reaches the fixed point.
