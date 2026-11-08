---
id: test-runner-core
title: Test runner in Narra for programs, panics and errors
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
  - selfhost/tests/**
  - scripts/test.sh
  - compiler/testdata/**/*.err
  - roadmaps/selfhost/nodes/test-runner-core/**
shared:
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: selfhost-driver
    kind: hard
    why: the runner builds samples through narrac build and run
---

# Test runner in Narra for programs, panics and errors

## Goal

A test runner written in Narra, `selfhost/tests/` (its own `narra.toml` and binary), with one entry script `scripts/test.sh`. It replaces the Go harness for: programs (81 plus stage2, debug with `-sanitize` and the leak check, and release; `.in` files fed to stdin), panics (9), compile errors (286, stage2 2, packages/errors 26), packages (8) and `runner/`. Later nodes add more suites, so each suite is its own module behind a small registry.

## Context

- Read `roadmaps/selfhost/nodes/seed-parity-spike/outcome.md` first: it has the measurements and the proposal this brief comes from.
- The driver's outcome (`roadmaps/selfhost/nodes/selfhost-driver/outcome.md`) lists the commands, flags and cc invocations.
- `Process.run` blocks, closes stdin, cannot set environment variables and has no timeout; std/fs has no remove or temp dir. Run each child through `sh -c` (environment, `< file`, `mkdir -p`, `rm -rf`). Parallelism: `scripts/test.sh` runs N shards of the runner (`--shard i/N`).
- The runner takes the testdata root as an argument, default `compiler/testdata`, so the later move to `testdata/` changes nothing.
- Decided: compile errors are compared in full against `.err` golden files. Generate each `.err` once from Go's `narra` output now, while Go exists (all 330 already match). Keep the `// error:` comments as documentation.
- Keep one sanitizer self-build: the runner builds the compiler with `-sanitize` and runs it on `selfhost/` (the fixed point), as `TestSelfhostBootstrap` does today.
- `narrac fmt` reads `narra.toml` through `$PWD`: set `PWD` for every child.

## Acceptance

- [ ] `sh scripts/test.sh` runs those suites on narrac and passes, with no `go` call, under the sanitizer and leak check.
- [ ] Breaking one `.out`, one `.err` and one panic sample makes it fail with a readable diff, then pass again when restored.
- [ ] Wall time with shards is reported in the outcome next to the Go suite's time for the same suites.

## Sizing

M: a new package; the suites are simple compare loops.

## Notes for the worker

- Do not edit `selfhost/src`. If the runner needs a compiler or std change, write it down as a change request in the node folder and work around it.
