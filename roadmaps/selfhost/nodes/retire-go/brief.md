---
id: retire-go
title: Delete the Go compiler and Go test harness
status: done
priority:
type: refactor
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
gate: true
added: 2026-11-08
spent: 192759
owns:
  - compiler/**
  - testdata/**
  - bench/run.sh
  - bench/README.md
  - scripts/**
  - selfhost/CLAUDE.md
  - selfhost/runtime/**
  - selfhost/src/runtimetext.nra
  - selfhost/src/stdtext.nra
  - selfhost/tests/**
  - README.md
  - docs/guide/01-getting-started.md
  - DECISIONS.md
  - roadmaps/selfhost/nodes/retire-go/**
shared: []
reserves: []
provides: []
depends_on:
  - on: checker-fixes
    kind: hard
    why: no diagnostic or fix may be lost when the Go checker goes
  - on: goldens-tools
    kind: hard
    why: every tool suite must pass on the Narra runner first
  - on: goldens-programs
    kind: hard
    why: the frozen goldens need Go to generate
  - on: seed-file
    kind: hard
    why: the seed replaces Go as the bootstrap
---

# Delete the Go compiler and Go test harness

## Goal

After this node, no `.go` file builds or tests the compiler. Go stays only as a benchmark language in `bench/`.

## Context

- Read `roadmaps/selfhost/nodes/seed-parity-spike/outcome.md` first: it has the measurements and the proposal this brief comes from.

## Steps

- [ ] Runtime source: move the C runtime (`compiler/internal/narra/runtime*.h` and the per-module `Module.Runtime` strings) to `selfhost/runtime/*.h`. Add a generator (`narrac gen-text` or a shell script) that writes `runtimetext.nra` and `stdtext.nra` from them, and a runner check that the generated files are current.
- [ ] Read the Go test table in `roadmaps/selfhost/nodes/goldens-programs/outcome.md`. Cover its 3 uncovered tests before deleting: the two generators (the runtime step above) and `TestEveryFormatHasCode`, by porting `roadmaps/selfhost/nodes/checker-fixes/check-codes.py` to a Narra runner suite. The test suite needs only `cc` and `sh`, no Python.
- [ ] Delete `compiler/cmd`, `compiler/internal` and `compiler/go.mod`.
- [ ] Switch `bench/run.sh`, `scripts/linux-check.sh`, `scripts/linux/Dockerfile`, `README.md`, `selfhost/CLAUDE.md` and `docs/guide/01-getting-started.md` to the seed and `scripts/test.sh`.
- [ ] Last: `git mv compiler/testdata testdata`, and update the runner's default root and the few non-roadmap files that name the old path.
- [ ] Very last: regenerate the seed with `sh scripts/seed-update.sh`, so `seed/narrac.c` matches the final source.

## Acceptance

- [ ] `sh selfhost/bootstrap.sh && sh scripts/test.sh` passes on macOS and in the Linux container, with `go` removed from the `PATH`.
- [ ] `git ls-files '*.go'` lists only files under `bench/`.
- [ ] Gate: the user sees the full suite result before the story pull request opens.

## Sizing

M: mostly deletion and path updates; the runtime move is the only new code.

## Notes for the worker

- Do not reword the roughly 282 comments in `selfhost/src` that mention Go. That is an optional later cleanup.

Landed on feat/c-seed as 9ca8464. Linux x86_64 was not run on real hardware: the maintainer has no x86_64 Linux host, and AddressSanitizer is killed under emulation. Accepted on 2026-11-09.
