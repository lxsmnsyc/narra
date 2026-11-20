---
id: compiler-std-files
title: "The compiler uses std file and folder functions"
status: todo
priority: should
type: refactor
story: batch-23
early: false
outline: false
milestone: p6-docs
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
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/compiler-std-files/**
shared:
  - selfhost/src/driver/**
  - testdata/**
  - scripts/**
reserves: []
provides: []
depends_on: []
---
# The compiler uses std file and folder functions

## Goal

Batch-22 added `Process.cwd`, `Fs.rename`, `Fs.copyFile`, `Fs.appendText` and `Fs.removeDirAll`. The compiler still runs shell commands for these:

- `driver/workspace.nra:75`: `sh -c "pwd -P"`. `Process.cwd` gives the current folder; check whether it resolves symbolic links as `pwd -P` does (the workspace code relies on the physical path). If it does not, keep the call and say so in the outcome.
- `driver/depcache.nra:47`: `runProgram("pwd", [])`.
- `driver/depcache.nra:333`: `rm -rf -- <work>/.git`, now `Fs.removeDirAll`.
- `driver/depcache.nra:337` and `:371`: `INSTALL_SCRIPT` and `REMOVE_SCRIPT` through `sh -c`. Replace them with `Fs` calls when the result is the same, including the atomic rename into the cache and the behavior when the target already exists (two builds filling the cache at once). Read the scripts and their tests first.
- `GIT_SCRIPT` stays: it sets `GIT_TERMINAL_PROMPT=0`, and `Process.run` cannot set environment variables (inbox).

## Rules

- Behavior does not change: the `deps`, workspace and LSP cases pass unchanged. Add a case only where a replaced script had a path no case covered.
- Errors keep their messages, or get clearer ones, and the cache stays consistent when a step fails.
- No function-level comments, no `<noun>Of` names, no D-numbers in code.

## Acceptance

- [ ] The calls above use std functions, except those kept with a reason.
- [ ] `sh scripts/test.sh` passes; `sh selfhost/bootstrap.sh` reaches the fixed point; `sh scripts/linux-check.sh` passes (the cache code runs on Linux too).
