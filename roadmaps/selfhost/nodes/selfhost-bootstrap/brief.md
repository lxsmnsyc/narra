---
id: selfhost-bootstrap
title: "Bootstrap: the compiler compiles itself"
status: active
priority: must
type: feature
story:
early: false
outline: false
milestone: m8-selfhost
target:
ship: pr
from:
size: M
spike: false
hypothesis:
verify: cd compiler && go test -count=1 ./internal/narra -run 'TestSelfhostBootstrap'
exclusive: []
fulfilled_by:
gate: true
added: 2026-11-04
spent:
owns:
  - compiler/internal/narra/selfhost_bootstrap_test.go
  - selfhost/bootstrap.sh
  - roadmaps/selfhost/nodes/selfhost-bootstrap/**
shared:
  - compiler/internal/narra/selfhost_test.go
  - compiler/internal/narra/selfhost_emit_test.go
  - compiler/README.md
  - selfhost/narra.toml
  - selfhost/src/**
reserves: []
provides: []
depends_on:
  - on: selfhost-emit
    kind: hard
    why: needs narrac to emit its own package byte for byte
---

# Bootstrap: the compiler compiles itself

## Goal

Close the loop. narrac only emits C (`narrac emit <path>`); the C file is self-contained and builds with `cc -std=c11 -w -o out out.c -lm`, the same command `Build` in `compiler/internal/narra/driver.go` runs.

- Stage 1: the Go compiler builds `selfhost/` (`narra build selfhost`), as `buildSelfhost` does now.
- Stage 2: stage 1 runs `narrac emit selfhost` and the C compiler builds that C.
- Fixed point: stage 2 runs `narrac emit selfhost`, and the result is byte-identical to stage 1's output (and to Go's `narra emit selfhost`).
- Stage 2 emits byte-identical C to Go on every emit input (`emitInputs` in `selfhost_emit_test.go`).

## Context

- Read the selfhost-emit outcome first: it gives the selfhost package input, the emit times and any divergences left.
- `buildSelfhost` in `selfhost_test.go` builds stage 1 once per test run with `modes[0].opts` (ASan and UBSan). Build stage 2 with the same flags so the stage-2 compiler is sanitizer-checked too. Cache it like `buildSelfhost` so it builds once per run.
- A stage-2 build costs one stage-1 emit of about 95k lines of C (0.19 s release, about 4 s under sanitizers) plus a C compile. Keep the whole test under about 60 s; if it is slower, report the time and gate it behind `testing.Short()`.

## Acceptance

- [ ] `TestSelfhostBootstrap` builds stage 2 from stage 1's C, checks the fixed point (stage-2 `emit selfhost` equals stage-1 output), and checks stage 2 against Go on every emit input, with leak checks on (`NARRA_CHECK_LEAKS=1`, as `runNarrac` sets).
- [ ] `selfhost/bootstrap.sh` runs the same steps by hand from a clean checkout (Go build, stage 2 build, fixed-point diff) and exits non-zero on any difference. It needs only `go` and `cc`.
- [ ] `compiler/README.md` has a "Bootstrapping" section: the steps, the commands, and what the fixed point proves.
- [ ] The outcome reports the times of stage 1 and stage 2 for `emit selfhost`, and the stage-2 binary size.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Out of scope

- A `narrac build` command that runs the C compiler. Add it as a new node if the user wants narrac to replace `narra build`.
- Removing the Go compiler (GOAL non-goal: keep it for bootstrapping).

## Sizing

M: one test file, one script, a README section. Divergences that only stage 2 shows (for example uninitialized memory that ASan in stage 1 did not catch) are fixed here if small; past M, split them into fix nodes and report.
