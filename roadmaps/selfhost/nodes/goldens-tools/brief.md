---
id: goldens-tools
title: Tool golden suites in the Narra test runner
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
  - roadmaps/selfhost/nodes/goldens-tools/**
shared: []
reserves: []
provides: []
depends_on:
  - on: test-runner-core
    kind: hard
    why: adds suites to the runner it builds
---

# Tool golden suites in the Narra test runner

## Goal

Port the tool suites to the Narra runner: dumps (astdump, declsdump, typesdump, pkgdump: 145 goldens), fmt (55 goldens, 10 configs, in-place, `--check`, `PWD` and directory cases, corpus idempotence), fmt-doc (14 plus errors), trivia (3 plus corpus `--check`), query (72), sources formatted (`TestSourcesFormatted`), node ids and the coverage greps. All already run on narrac only.

## Context

- Read `roadmaps/selfhost/nodes/seed-parity-spike/outcome.md` first: it has the measurements and the proposal this brief comes from.
- The runner's outcome (`roadmaps/selfhost/nodes/test-runner-core/outcome.md`) describes the suite registry.

## Acceptance

- [ ] Each suite passes on narrac through `sh scripts/test.sh`, and fails when one golden is broken.

## Sizing

M: mechanical, one module per suite.
