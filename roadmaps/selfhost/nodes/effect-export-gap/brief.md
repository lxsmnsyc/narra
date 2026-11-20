---
id: effect-export-gap
title: "Reproduce the exported uses alloc printing gap"
status: active
priority: could
type: fix
story: batch-20
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
added: 2026-11-19
spent:
owns:
  - roadmaps/selfhost/nodes/effect-export-gap/**
shared:
  - selfhost/src/check/**
  - testdata/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on: []
---
# Reproduce the exported `uses alloc` printing gap

## Goal

Inbox, 2026-11-15, from narra-skills: "an exported function declared `uses alloc` that prints compiled with no `uses io`." The orchestrator could not reproduce it on main (2026-11-19): a single file with `function noisy() uses alloc { print("hi") }`, and a library file `lib/noisy.nra` with `export function noisy(): string uses alloc`, both report "uses `io` (through `print`) but its signature does not declare it".

Look for the shape the narra-skills node met: its outcome and log (`roadmaps/selfhost/nodes/narra-skills/`), the skill examples under `skills/` or wherever narra-skills put them, a package with `narra.toml` and `exports`, `narrac build` versus `check`, a function used only through a function value or a trait. If none shows the gap, write "Not reproduced" with what you tried and stop. If one does, fix it with a failing case first. This is separate from the trait-implementation gap in the inbox (an implementation's effects not limited by its trait function); do not take that one on.

## Acceptance

- [ ] Either "Not reproduced" with the shapes tried, or a case that fails on the batch base and passes on the branch.
- [ ] `sh scripts/test.sh` passes.
