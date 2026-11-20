---
id: std-async-fold
title: "Fold std/async/* into std/fs, std/time and std/net"
status: active
priority: could
type: feature
story: batch-22
early: false
outline: false
milestone: p6-docs
target:
ship: auto
land: auto
from:
size: L
spike: false
hypothesis:
verify: sh scripts/test.sh
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-20
spent:
owns:
  - roadmaps/selfhost/nodes/std-async-fold/**
shared:
  - std/**
  - selfhost/**
  - testdata/**
  - SPEC.md
  - DECISIONS.md
  - compiler/README.md
  - docs/**
  - skills/narra/**
  - scripts/**
reserves: []
provides: []
depends_on: []
---
# Fold `std/async/*` into `std/fs`, `std/time` and `std/net`

## Goal

Apply the batch-22 decisions (see `nodes/batch-22/brief.md`): remove `std/async/file`, `std/async/time` and `std/async/net`.

- `std/net`: the module `std/async/net` under its new path, names unchanged (`Listener`, `Connection`, `listen`, `localPort`, `accept`, `connect`, `read`, `chunks`, `write`).
- `Time.sleep` in `std/time` becomes async (what `AsyncTime.sleep` does today, including its `Result` and cancel behavior). The blocking sleep goes. Programs that used the blocking `Time.sleep` (for example `testdata/channel_full_blocks.nra`) move to the async one or to another way to wait.
- `std/async/file` goes; its callers use `std/fs`.
- Everything that names the old modules follows: the registry and runtime, `std/src`, `testdata` and the test suites, `SPEC.md` (sections 8 and the std list), `compiler/README.md`, `DECISIONS.md` gets a new decision record for the rule (sync when the OS cannot wait in the background, async when it can; one form per API), the guide (chapters 8, 11, 13, 14, 15, 20 and the index), `skills/narra/`, `scripts/cross-check.sh`, `scripts/std/generate.nra`'s module table, and the generated std reference (regenerate it).
- No migration errors or shims for the old paths (Narra is unreleased): an import of `std/async/file` is an unknown module like any other.

## Acceptance

- [ ] No `std/async/` path remains outside history and roadmap records; `std/net` exists; `Time.sleep` is async; `std/fs` is the only file API.
- [ ] The guide teaches the new forms; the std reference is regenerated; SPEC and DECISIONS record the rule.
- [ ] `sh scripts/test.sh` passes; `pnpm --dir website build` passes.
