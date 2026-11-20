---
id: entry-args-unused
title: "No unused warning for the entry function's parameters"
status: active
priority: could
type: fix
story: batch-22
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
  - roadmaps/selfhost/nodes/entry-args-unused/**
shared:
  - selfhost/src/check/**
  - testdata/**
  - selfhost/tests/src/**
reserves: []
provides: []
depends_on: []
---
# No unused warning for the entry function's parameters

## Goal

`export function run(args: Array<string>): int` is the fixed entry signature, so the "is never read" warning on `args` cannot be acted on without renaming it to `_args`. Most guide samples print it. Do not report the entry function's parameters as unused (plain and `async` entry, in a single file and in a package's binaries). Prove it first: a sample whose only warning is about `args` prints it on the batch base.

## Acceptance

- [ ] A case that warns on the base and is clean on the branch; other unused parameters still warn.
- [ ] `sh scripts/test.sh` passes.
