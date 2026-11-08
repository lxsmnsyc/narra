---
id: selfhost-driver
title: build, run, check and test commands in narrac
status: done
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
spent: 133269
owns:
  - selfhost/src/main.nra
  - selfhost/src/driver*.nra
  - roadmaps/selfhost/nodes/selfhost-driver/**
shared:
  - compiler/README.md
reserves: []
provides: []
depends_on: []
---

# build, run, check and test commands in narrac

## Goal

Give the self-hosted compiler the user commands that today only the Go `narra` has: `narrac build`, `narrac run`, `narrac check` and `narrac test`, with the same flags and the same `cc` invocations per mode (debug with AddressSanitizer and UBSan, release with `-O2 -DNARRA_RELEASE`, the leak check through `NARRA_CHECK_LEAKS=1`).

## Context

- The Go driver is `compiler/cmd/narra` and `BuildOptions` in `compiler/internal/narra`. The sanitizer flags per mode are in `run_test.go` (`modes`).
- The runtime can already start processes (`runtime_io.h`, the `execvp` path). Use the std process API to call `cc`.
- `narrac` today has `emit`, `fmt`, `query`, `parse`, `pkg`, `decls`, `types` and dump commands (`selfhost/src/main.nra`).

## Acceptance

- [ ] `narrac build`, `run`, `check` and `test` accept the same arguments as Go's `narra` and give the same exit codes and output on `compiler/testdata` samples (spot-check 10 programs, 10 error samples, the `runner/` samples).
- [ ] The C compiler is `$CC` when set, else `cc`.
- [ ] Fast check passes: `cd compiler && go test ./internal/narra -run 'TestSelfhostBootstrap|TestStage2|TestSourcesFormatted'`.

## Sizing

M: one new module with argument parsing and process calls; the compiling itself already exists.

## Notes for the worker

- Put the driver in its own module (`selfhost/src/driver.nra`) so `main.nra` only dispatches.
- Keep the Go compiler working; it is deleted by retire-go.

Landed on feat/c-seed as de4096e.
