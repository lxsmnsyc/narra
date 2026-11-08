---
id: goldens-programs
title: Guide, std and frozen edge cases in the Narra runner
status: todo
priority:
type: feature
story: c-seed
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: 
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-08
spent:
owns:
  - selfhost/tests/**
  - compiler/testdata/lex/**
  - compiler/testdata/parse/**
  - compiler/testdata/habits/**
  - compiler/testdata/manifest/**
  - compiler/testdata/std-tests/**
  - roadmaps/selfhost/nodes/goldens-programs/**
shared: []
reserves: []
provides: []
depends_on:
  - on: test-runner-core
    kind: hard
    why: adds suites to the runner it builds
  - on: checker-fixes
    kind: hard
    why: the habit cases check fixes, so narrac must produce them
---

# Guide, std and frozen edge cases in the Narra runner

## Goal

Port the remaining language suites: guide samples (177 plus the extractor), runner tests (8, environment variables and temp dirs), std test failures (9 inline Go programs become testdata files), cnames properties (3) and the 2 emit-error messages. Freeze every case that lives only inside Go test files into testdata files with goldens made from Go now: 16 lex and 82 parse edge cases, 20 habit cases, 8 manifest errors.

## Context

- Read `roadmaps/selfhost/nodes/seed-parity-spike/outcome.md` first: it has the measurements and the proposal this brief comes from.
- This must land before retire-go: once Go is gone, these goldens cannot be generated.

## Acceptance

- [ ] Each suite passes on narrac through `sh scripts/test.sh`, and fails when one golden is broken.
- [ ] Every inline case listed in the spike exists as a testdata file with a golden.

## Sizing

M: mostly moving cases into files and generating goldens once.
