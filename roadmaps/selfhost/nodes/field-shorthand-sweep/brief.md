---
id: field-shorthand-sweep
title: "Use field shorthand in the compiler and std"
status: todo
priority: must
type: refactor
story: batch-15
early: false
outline: false
milestone: p3-review
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
added: 2026-11-14
spent:
owns:
  - roadmaps/selfhost/nodes/field-shorthand-sweep/**
shared:
  - selfhost/src/*.nra
  - std/src/*.nra
  - seed/narrac.c
reserves: []
provides: []
depends_on:
  - on: field-shorthand
    kind: hard
    why: the sweep uses the feature
  - on: remove-lazy
    kind: hard
    why: the sweep touches every compiler file; it runs after the other compiler children land
  - on: coroutine-send
    kind: hard
    why: runs after the other compiler children
  - on: range-step
    kind: hard
    why: runs after the other compiler children
  - on: float-shortest
    kind: hard
    why: runs after the other compiler children
  - on: slot-shared-writes
    kind: hard
    why: runs after the other compiler children
  - on: index-assign-order
    kind: hard
    why: runs after the other compiler children
  - on: stream-exit
    kind: hard
    why: runs after the other compiler children
  - on: generator-option-params
    kind: hard
    why: runs after the other compiler children
  - on: batch-15-nits
    kind: hard
    why: runs after the other compiler children
  - on: remove-inline
    kind: hard
    why: runs after the other compiler children
  - on: const-functions-std
    kind: hard
    why: runs after the other compiler children
---
# Use field shorthand in the compiler and std

## Goal

After the batch seed update (the orchestrator runs it before this node starts), rewrite every `name: name` field in `selfhost/src` and `std/src` to the shorthand. About 512 sites in the compiler.

## Context

- Mechanical: do it with a script over the AST positions (`narrac query` or a small Narra or Python pass), then `narrac fmt --check`. Do not hand-edit.
- Regenerate `stdtext.nra` with `sh scripts/test.sh --suite generated --update`, then the seed twice (`sh scripts/seed-update.sh`).
- The emitted C for the compiler must be byte-identical before and after (shorthand is syntax only): compare `narrac emit selfhost` output.

## Acceptance

- [ ] No `name: name` field left in `selfhost/src` and `std/src` (a grep in `log.md`).
- [ ] `emit selfhost` C identical before and after; `sh selfhost/bootstrap.sh` reports the seed current.
- [ ] `sh scripts/test.sh` passes.

`selfhost/src` and seed-compiled std may use only features the committed seed compiles. Proof first: show the failure on the batch base before changing code, and record command and result in `log.md`. Scratch folder: /private/tmp/claude-501/<node id>/. Follow `roadmaps/selfhost/CLAUDE.md` and the ts-style rules there (no function-level comments, no D-numbers in code).
