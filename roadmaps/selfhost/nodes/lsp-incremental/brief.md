---
id: lsp-incremental
title: "narrac lsp: recheck only changed modules"
status: todo
priority: should
type: feature
story:
early: false
outline: false
milestone: m9-adoption
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostLsp|TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-07
spent:
owns:
  - selfhost/src/lspcache.nra
  - compiler/testdata/lsp/incremental_*/**
  - roadmaps/selfhost/nodes/lsp-incremental/**
shared:
  - selfhost/src/lsp.nra
  - selfhost/src/modules.nra
  - selfhost/src/body.nra
  - selfhost/src/check.nra
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: language-service
    kind: hard
    why: replaces the server's whole-package recheck with a cached one
---

# narrac lsp: recheck only changed modules

## Goal

After a change, the server parses and checks only the changed module and the modules that import it, directly or not, and reuses the results of the rest. SPEC section 4 "Inference": other modules check against exported signatures only, so an edit that keeps a module's exported signatures leaves its importers' results valid.

## Context

- Read `nodes/language-service/outcome.md`. The self-hosted checker keeps results in package-wide side tables indexed by integer node ids (D63, D70), and node ids run across the package in load order (`narrac pkg --ids`). Reuse needs ids that stay stable for unchanged modules, for example a per-module id range, or a cache keyed by module that is rebuilt from its own results.
- Measure on `selfhost/` itself opened as a package: the time from `didChange` of one leaf module to `publishDiagnostics`, before and after.

## Acceptance

- [ ] An edit inside a function body of one module rechecks that module only; an edit to an exported signature also rechecks its importers; diagnostics are the same as a full recheck (a test compares both paths on scripted edits in `compiler/testdata/lsp/incremental_*/`).
- [ ] On `selfhost/` as a package, a body edit in a leaf module answers at least 3x faster than the full recheck of language-service, measured as in perf-pass (`bench.py`), and the outcome reports both times.
- [ ] `narrac check`, `narrac query` and the bootstrap output stay unchanged.
- [ ] Fast check passes.

## Sizing

M. Uncertainty medium (id stability across modules in the checker's side tables), blast radius medium (loader and checker entry points), verifiable by the comparison test and the benchmark.

## Notes for the worker

- If stable ids need a change to the checker's tables that is larger than this node, stop after the measurement and write the design in the outcome; the orchestrator splits the work.
