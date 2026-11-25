---
id: process-run-options
title: "Process.run options: input, environment, working folder, inherited output"
status: done
priority: should
type: feature
story: batch-24
early: false
outline: false
milestone: p2-std
target:
ship: auto
land: auto
from:
size: M
risk: medium
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-25
spent: 383300
owns:
  - roadmaps/selfhost/nodes/process-run-options/**
  - testdata/process_run*
shared:
  - selfhost/runtime/*
  - selfhost/src/runtime*.nra
  - selfhost/src/std/stdtext.nra
  - docs/reference/std/README.md
  - docs/reference/std/*.md
  - SPEC.md
reserves: []
provides: []
depends_on: []
---

# Process.run options: input, environment, working folder, inherited output

## Goal

`Process.run` takes options: text for standard input, environment variables to add or remove, a working folder, and whether output is captured or passed through to the parent. Calls without options keep working unchanged, through parameter defaults.

## Context

- Source: std-gap-survey rank 9: `driver/driver.nra` `startProgram` feeds stdin through `sh -c 'exec "$@" < file'`; `driver/depcache.nra` sets `GIT_TERMINAL_PROMPT` through `sh -c`. `inbox.md` 2026-11-15 (deps-fetch).
- Today: `run(command: string, arguments: Array<string>): Result<#[int, string, string], string>` (`docs/reference/std/process.md`).
- Writing stdin while reading stdout and stderr must not deadlock on a full pipe: use `poll` (or threads on Windows), not write-then-read.
- The `pkg-changesets` node runs git through this. Replacing the compiler's `sh -c` calls waits for a seed that knows the options.

## Acceptance

- [x] A sample sends 1 MiB on stdin to `cat` and reads it back whole.
- [x] A sample runs `env` with one variable added and one removed, and `pwd` in a given folder, and checks the output.
- [x] Passed-through output appears on the parent's stdout in order; captured output is then empty.
- [x] Existing calls compile and behave as before (`sh scripts/test.sh` passes).

## Sizing

M: POSIX `posix_spawn` file actions and the Windows `CreateProcess` path, plus the poll loop.
