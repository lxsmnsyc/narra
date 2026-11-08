---
id: retire-go
title: Delete the Go compiler and Go test harness
status: todo
priority:
type: refactor
story: c-seed
early: false
outline: true
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
spent:
owns: []
shared: []
reserves: []
provides: []
depends_on:
  - on: checker-parity
    kind: hard
    why: no diagnostic may be lost when the Go checker goes
  - on: test-runner-goldens
    kind: hard
    why: every ported suite must pass on the Narra runner first
  - on: seed-file
    kind: hard
    why: the seed replaces Go as the bootstrap
---

# Delete the Go compiler and Go test harness

## Goal

Remove `compiler/cmd/narra`, `compiler/internal/narra` and `compiler/go.mod`. Keep `compiler/testdata` (or move it, as the spike proposes). Switch `bench/run.sh`, `scripts/linux-check.sh`, the READMEs and every `CLAUDE.md` outside the roadmap to the seed and the Narra runner. Go remains only as a benchmark language in `bench/`. Gated: the user checks the full suite result before this lands. Outline until the spike reports.
