---
id: entry-args-unused
title: "No unused warning for the entry function's parameters"
status: done
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

Resolved 2026-11-20 (orchestrator): the compiler already skips the entry's parameters (`entryParameter` in `check.nra`, since 2026-11-07); the CLI does not reproduce it. The warnings the maintainer saw are on the website: the twoslash plugin wrapped each sample in a package whose `narra.toml` declares no binary, so `run` was not an entry and 63 samples showed `"args" is never read` inline. `website/plugins/twoslash.ts` now declares `[[binary]] entry = "main.run"` for samples that export `run`; the built site has no `args` warning left.
