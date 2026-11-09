---
id: style-sweep
title: "Code style sweep: step comments, no D-numbers, action names"
status: todo
priority: should
type: refactor
story: batch-2
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-09
spent:
owns:
  - selfhost/src/**
  - std/src/**
  - selfhost/tests/src/**
  - roadmaps/selfhost/nodes/style-sweep/**
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
  - on: template-nested-strings
    kind: hard
    why: the sweep touches every hot file, so it runs after the rest of the batch
  - on: effect-variable-values
    kind: hard
    why: same files
  - on: generator-function-type
    kind: hard
    why: same files
  - on: misleading-hints
    kind: hard
    why: same files
  - on: c-style-for
    kind: hard
    why: same files
  - on: compound-assign
    kind: hard
    why: same files
  - on: cell-capture
    kind: hard
    why: same files
  - on: std-basics
    kind: hard
    why: same files
  - on: spec-samples-test
    kind: hard
    why: same files
---

# Code style sweep: step comments, no D-numbers, action names

## Goal

User request, 2026-11-09. The code style leaks past ts-style: batch-1 added 275 functions with a Go-style comment on the function itself (`// checkModules is ...`), comments citing D-numbers, and `<noun>Of` names (`keyOf`, `keysOf`, `prefixOf` in `types.nra` and `stdsized.nra`). Older selfhost code has the same comments from the Go port.

Sweep `selfhost/src`, `std/src` and `selfhost/tests/src`:

- Move what a comment on a function explains into short comments on the steps inside it, or drop it when the name already says it.
- Drop D-number citations from code comments. Keep the rule the comment states.
- Rename `<noun>Of` functions to action names (`keyOf` to `findKey` or `buildKey`, whatever the function does), with every caller.
- No behavior change. Do not use features the committed seed cannot compile.

Work file by file, one commit per group of files, so a failure is easy to bisect.

## Acceptance

- [ ] `grep -rnE '^\s*// [a-z][A-Za-z]* (is|returns|reports|checks) ' selfhost/src std/src selfhost/tests/src` finds no function-level comments of the Go form (list any kept on purpose in the outcome).
- [ ] `grep -rnE 'D[0-9]{2}' selfhost/src std/src selfhost/tests/src` finds no D-number in a comment.
- [ ] No function name ends in `Of`, except names that are not `<noun>Of` (list them).
- [ ] Every sample's emitted C is unchanged (the emit goldens pass without updates), and `sh selfhost/bootstrap.sh` reaches the fixed point.
- [ ] `sh scripts/test.sh` passes, including the `sources` format check.
