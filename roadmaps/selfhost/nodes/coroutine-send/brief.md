---
id: coroutine-send
title: "Co.send for two-way generators"
status: done
priority: must
type: feature
story: batch-15
early: false
outline: false
milestone: p3-review
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
added: 2026-11-14
spent: 111837
owns:
  - roadmaps/selfhost/nodes/coroutine-send/**
  - testdata/coroutine*
  - testdata/errors/coroutine*
shared:
  - std/src/coroutine.nra
  - selfhost/src/stdtext.nra
  - selfhost/src/calls.nra
  - selfhost/src/stmts.nra
  - selfhost/src/scopes.nra
  - selfhost/src/effects.nra
  - selfhost/src/stdinternal.nra
  - docs/guide/*.md
  - selfhost/src/diagnostics.nra
  - SPEC.md
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---
# Co.send for two-way generators

## Goal

D119: the two-argument `Co.resume(mutable co, value)` becomes `Co.send(mutable co, value)`; one-argument `Co.resume` stays. No function is picked by argument count. No deprecation path: old two-argument calls get the usual wrong-argument-count error.

## Context

- `std/src/coroutine.nra` and the checker rows that know `resume` (`grep -rn resume selfhost/src`). Async generators (D110) may share the code path: check `Async.next` does not depend on the name.
- Update SPEC, the guide's generators section, samples and goldens.

## Acceptance

- [ ] Coroutine samples use `Co.send`; a test that the two-argument `Co.resume` is now an error.
- [ ] `sh scripts/test.sh` passes; stdtext regenerated through the generated suite.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
