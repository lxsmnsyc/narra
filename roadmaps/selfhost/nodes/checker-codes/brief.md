---
id: checker-codes
title: Diagnostic codes, check --json and explain in narrac
status: active
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
  - selfhost/src/check.nra
  - selfhost/src/parser.nra
  - selfhost/src/lexer.nra
  - selfhost/src/modules.nra
  - selfhost/src/manifest.nra
  - selfhost/src/diagnostics.nra
  - selfhost/src/queryjson.nra
  - selfhost/src/driver.nra
  - selfhost/src/*.nra (report call sites only)
  - roadmaps/selfhost/nodes/checker-codes/**
shared:
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-driver
    kind: hard
    why: adds --json and explain to the driver commands
---

# Diagnostic codes, check --json and explain in narrac

## Goal

Every diagnostic the self-hosted compiler reports carries the same code as Go's. `narrac check --json` prints diagnostics per `contracts/diagnostic.md`, and `narrac explain <code>` prints the explanation, both matching Go's `narra`.

## Context

- Read `roadmaps/selfhost/nodes/seed-parity-spike/outcome.md` first: it has the measurements and the proposal this brief comes from.
- Text and positions already match Go on all 330 error samples. Only codes, JSON and `explain` are missing.
- Codes cannot be derived from the message text (113 of 410 are ambiguous). Pass the code at each report site: about 250 checker calls, plus parser, lexer and loader sites.
- Port the 289-row table of `compiler/internal/narra/diagnostic_codes.go` as data into `selfhost/src/diagnostics.nra`.
- Replace the `E0000` placeholder in `queryjson.nra`.

## Acceptance

- [ ] On every sample in `compiler/testdata/errors/`, `diagnostics/`, `packages/errors/` and `stage2/errors/`, `narrac check --json` output equals Go's `narra check --json`, byte for byte. A throwaway comparison script in the node folder is fine.
- [ ] `narrac explain` equals Go for every code in the table.
- [ ] A check (test or script) fails when a report site passes no code.
- [ ] Fast check passes: `cd compiler && go test ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`.

## Sizing

M: mechanical, many call sites, one data table.

## Notes for the worker

- Fix suggestions are the next node (checker-fixes). Leave a `fixes` field empty in the JSON shape if the contract has one.
