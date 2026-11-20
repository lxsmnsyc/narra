---
id: std-fs-more
title: "Append, rename, copy, remove a folder tree, current folder"
status: todo
priority: could
type: feature
story: batch-22
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-fs-more/**
shared:
  - selfhost/src/std/**
  - selfhost/runtime/**
  - selfhost/src/runtime/**
  - std/src/**
  - testdata/**
  - selfhost/tests/src/**
  - docs/**
  - scripts/**
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: std-async-fold
    kind: hard
    why: edits std/fs and the files chapter
---
# Append, rename, copy, remove a folder tree, current folder

## Goal

Survey item 8: `Fs.appendText`, `Fs.rename`, `Fs.copyFile`, `Fs.removeDirAll`, `Process.cwd` (runtime). Chapter 11 drops "There is no function that appends" and teaches them. List the `sh -c` workarounds in `selfhost/src` (depcache `mv` and `rm -rf`, workspace `pwd -P`) in the outcome for the follow-up after the seed is regenerated.

## Rules for every std addition

- Read `roadmaps/selfhost/nodes/std-gap-survey/survey.md` for the evidence of each piece.
- Write it in Narra in `std/src` where it can be; use the runtime only where the survey says so.
- Each export gets a registry row (or source declaration) with a readable display signature (see std-registry-signatures' outcome); the `query/builtin_signatures` case must pass.
- Replace the hand-written copies in `selfhost/src`, `scripts/` and `testdata/` that the survey lists, where the seed rule allows (`selfhost/src` may use only what the committed seed compiles: a new std function is usable in `selfhost/src` only after the seed is regenerated, so leave compiler call sites for a follow-up and list them in the outcome).
- Teach it in the guide where the survey's evidence points (one short runnable sample each) and regenerate the std reference (`sh scripts/test.sh --suite generated --update`).

## Acceptance

- [ ] Each function exists with a case per function (and the edge cases the survey names), a display signature, a guide sample, and a regenerated std reference page.
- [ ] `sh scripts/test.sh` passes; `pnpm --dir website build` passes.
