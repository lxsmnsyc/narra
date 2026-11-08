---
id: seed-parity-spike
title: Measure what only the Go compiler and Go tests cover
status: todo
priority:
type: spike
story: c-seed
early: false
outline: false
milestone:
target:
ship: auto
from:
size: M
spike: true
hypothesis:
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-08
spent:
owns:
  - roadmaps/selfhost/nodes/seed-parity-spike/**
shared: []
reserves: []
provides: []
depends_on: []
---

# Measure what only the Go compiler and Go tests cover

## Goal

List everything that would be lost if `compiler/` (Go) were deleted today, so the other children of story c-seed are sized from facts. This node writes a report and changes no code.

## Steps

- [ ] For each Go test file in `compiler/internal/narra/*_test.go`, say what it checks, which testdata suite it reads, and whether it compares Go against the self-hosted compiler (goes away with Go) or checks the language (must be ported to the Narra runner).
- [ ] Run the self-hosted compiler (stage 2, built by `selfhost/bootstrap.sh`) on every file in `compiler/testdata/errors/` and `compiler/testdata/diagnostics/`, and compare its diagnostics (code, message, position) with the expected output. Count matches and list each mismatch by file.
- [ ] Do the same for `panics/`, `fmt/`, `query/`, `packages/`, `runner/` and the guide samples (`docs/guide`, TestGuide), using whatever command the self-hosted compiler has.
- [ ] List the `narra` (Go) commands and flags the self-hosted `narrac` lacks: `build`, `run`, `check`, `test`, `lex`, `parse`, `-release`, sanitizer and leak modes, and how Go calls `cc` (flags per mode, from `run_test.go` and `cmd/narra`).
- [ ] List every file outside `compiler/` that runs the Go compiler (`bench/run.sh`, `scripts/linux-check.sh`, `selfhost/bootstrap.sh`, docs).

## Acceptance

- [ ] `outcome.md` has: the test-file table, the diagnostic mismatch count and list, the missing-command list, and the outside-file list.
- [ ] `outcome.md` proposes a size for checker-parity, test-runner-core and test-runner-goldens, and says whether any should split.

## Sizing

M: reading 34 test files and running about 300 error samples through stage 2, no code.

## Notes for the worker

- Do not fix anything. Mismatches are input for checker-parity.
- A small throwaway script in the node folder to diff diagnostics is fine.
